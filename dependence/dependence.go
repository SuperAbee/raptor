package dependence

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"raptor/configcenter"
	"raptor/constants"
	"raptor/filter"
	"raptor/proto"
	"sync"
	"time"
)

var configCenter = configcenter.New()
var locks [10]*sync.Mutex

func init() {
	for i := 0; i < len(locks); i++ {
		locks[i] = new(sync.Mutex)
	}
}

func ExecuteSingleTask(jobInstance *proto.JobInstance) {
	jobInstance.Config.InstanceID = jobInstance.ID
	fmt.Println("任务 %s 开始执行 时间%s\n", jobInstance.ID, time.Now().String())
	go filter.NewChain(*jobInstance).Do()
}

func test(instanceId, mainId, id string) {
	fmt.Printf("任务 %s 开始执行 时间%s\n", id, time.Now().String())
	time.Sleep(10 * time.Second)
	fmt.Printf("任务 %s 执行结束\n", id)
	afterSingleComplete(instanceId, mainId, id, constants.TASK_COMPLETED)
}

func ExecuteDependenceTask(jobInstance *proto.JobInstance) (bool, error) {
	config := &jobInstance.Config
	fmt.Printf("依赖任务---%s---开始执行\n", config.ID)
	index := crc32.ChecksumIEEE([]byte(jobInstance.ID)) % 10
	locks[index].Lock()
	defer locks[index].Lock()
	findAndExcute(jobInstance.ID, config.ID, &config.Dependencies)
	//获取到任务后，更新任务状态
	_, err := saveDepenConfig(config.ID, config.Dependencies)
	if err != nil {
		log.Println(err)
		return false, err
	}
	return true, nil
}

//获取可以执行的任务
func getAvaiTask(dependencies proto.Dependency) []string {
	predecessor := make(map[string][]string)
	for _, link := range dependencies.Links {
		//如果前置任务已完成，直接跳过
		if dependencies.Nodes[link.From] == constants.TASK_COMPLETED {
			continue
		}
		preTasks, ok := predecessor[link.To]
		if !ok {
			preTasks = []string{}
		}
		preTasks = append(preTasks, link.From)
		predecessor[link.To] = preTasks
	}
	availTasks := []string{}
	//获取所有前置任务已完成的id
	for k, v := range dependencies.Nodes {
		if v != constants.TASK_UNEXECUTED {
			continue
		}
		m, ok := predecessor[k]
		if !ok {
			availTasks = append(availTasks, k)
			continue
		}
		if len(m) == 0 {
			availTasks = append(availTasks, k)
			continue
		}
	}
	return availTasks
}

func afterSingleComplete(jobInstanceID, mainID string, ID string, status string) {
	if mainID == "" {
		return
	}
	index := crc32.ChecksumIEEE([]byte(jobInstanceID)) % 10
	locks[index].Lock()
	defer locks[index].Unlock()
	depenConfig, err := getDepenConfig(mainID)
	if err != nil {
		log.Println(err)
		return
	}
	depenConfig.Nodes[ID] = status
	findAndExcute(jobInstanceID, mainID, &depenConfig)
	_, err2 := saveDepenConfig(mainID, depenConfig)
	if err2 != nil {
		log.Println(err2)
	}
}

func findAndExcute(jobInstanceID, mainID string, depenConfig *proto.Dependency) {
	tasks := getAvaiTask(*depenConfig)
	//如果没有获取到可以执行的任务
	if len(tasks) == 0 {
		allEnd := true
		//判断是否所有子任务都已经执行完毕
		for _, task := range depenConfig.Nodes {
			if task != constants.TASK_COMPLETED {
				allEnd = false
				break
			}
		}
		if allEnd {
			fmt.Printf("依赖任务---%s---已完成\n", mainID)
		}
	}
	for _, task := range tasks {
		//任务执行时更新状态
		depenConfig.Nodes[task] = constants.TASK_RUNNING
		config, err := configCenter.Get(task)
		if err != nil {
			depenConfig.Nodes[task] = constants.TASK_UNEXECUTED
			log.Fatal(err)
			continue
		}
		taskConfig := proto.Config{}
		err = json.Unmarshal([]byte(config.Content), &taskConfig)
		if err != nil {
			depenConfig.Nodes[task] = constants.TASK_UNEXECUTED
			log.Println(err)
			continue
		}
		taskConfig.InstanceID = jobInstanceID
		//go test(jobInstanceID, mainID, taskConfig.ID)
		go filter.NewChain(proto.JobInstance{Config: taskConfig})
	}
}

func getDepenConfig(ID string) (proto.Dependency, error) {
	config, err := configCenter.Get(ID + "Dependence")
	if err != nil {
		return proto.Dependency{}, err
	}
	depen := proto.Dependency{}
	err = json.Unmarshal([]byte(config.Content), &depen)
	if err != nil {
		return proto.Dependency{}, err
	}
	return depen, nil
}

func saveDepenConfig(mainID string, configDepen proto.Dependency) (bool, error) {
	depen, err := json.Marshal(configDepen)
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	return configCenter.Save(configcenter.Config{ID: mainID + "Dependence", Content: string(depen)})
}
