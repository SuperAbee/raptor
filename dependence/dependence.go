package dependence

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"raptor/configcenter"
	"raptor/constants"
	"raptor/eventcenter"
	"raptor/filter"
	"raptor/proto"
	"raptor/uuid"
	"strconv"
	"sync"
	"time"
)

var configCenter = configcenter.New()
var eventCenter = eventcenter.New()
var locks [10]*sync.Mutex
var sf *uuid.SnowFlakeUUID

func init() {
	for i := 0; i < len(locks); i++ {
		locks[i] = new(sync.Mutex)
	}
	eventCenter.Subscribe("job complete", func(event *eventcenter.Event) {
		data := event.Header
		instacne, _ := event.Body.(proto.JobInstance)
		fmt.Println("event get")
		AfterSingleComplete(instacne, data["status"])
	})
	sf, _ = uuid.NewSnowFlakeUUID((time.Now().Unix() % 1024) + 1)
}

func ExecuteSingleTask(jobInstance *proto.JobInstance) {
	fmt.Printf("任务 %s 开始执行 时间%s\n", jobInstance.ID, time.Now().String())
	jobInstance.StartTime = time.Now().Unix()
	saveInstanceConfig(*jobInstance)
	//向事件中心发送通知任务开始执行
	event := eventcenter.NewEvent().WithBody(*jobInstance)
	eventCenter.Publish("job start", event)
	//test(jobInstance)
	go filter.NewChain(*jobInstance).Do()
}

func test(instance *proto.JobInstance) {
	event := eventcenter.NewEvent().WithBody(*instance).WithHeader("status", constants.TASK_COMPLETED)
	fmt.Println("job execute")
	eventCenter.Publish("job complete", event)
}

func ExecuteDependenceTask(jobInstance *proto.JobInstance) (bool, error) {
	config := &jobInstance.Config
	fmt.Printf("依赖任务---%s---开始执行\n", config.ID)
	//向事件中心发送通知任务开始执行
	event := eventcenter.NewEvent().WithBody(*jobInstance)
	eventCenter.Publish("job start", event)
	jobInstance.StartTime = time.Now().Unix()
	// 保存实例信息
	saveInstanceConfig(*jobInstance)
	index := crc32.ChecksumIEEE([]byte(jobInstance.ID)) % 10
	fmt.Printf("lockB lock %v \n", index)
	locks[index].Lock()
	findAndExcute(jobInstance.ID, &config.Dependencies)
	//获取到任务后，更新任务状态
	// 保存引来信息
	_, err := saveDepenConfig(jobInstance.ID, config.Dependencies)
	if err != nil {
		log.Println(err)
		return false, err
	}
	fmt.Printf("lockB Unlock %v \n", index)
	locks[index].Unlock()
	return true, nil
}

//获取可以执行的任务
func getAvaiTask(dependencies proto.Dependency) []string {
	predecessor := make(map[string][]string)
	for _, link := range dependencies.Links {
		//如果前置任务已完成，直接跳过
		if dependencies.Nodes[link.From].Status == constants.TASK_COMPLETED {
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
		if v.Status != constants.TASK_UNEXECUTED {
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

func AfterSingleComplete(jobInstance proto.JobInstance, status string) {
	//更新此任务的实例状态
	jobInstance.Status = status
	jobInstance.EndTime = time.Now().Unix()
	saveInstanceConfig(jobInstance)
	//向事件中心发送通知任务结束
	event := eventcenter.NewEvent().WithBody(jobInstance)
	eventCenter.Publish("job end", event)
	fmt.Printf("任务 %v 完成", jobInstance.ID)
	//额外判断 是依赖任务的子任务实例
	if jobInstance.Type == constants.DEPENDENCE_SUB_JOB {
		//获取到整体依赖任务的实例id
		index := crc32.ChecksumIEEE([]byte(jobInstance.PreId)) % 10
		fmt.Printf("lockA lock %v \n", index)
		locks[index].Lock()
		fmt.Println(jobInstance.PreId)
		depenConfig, err := getDepenConfig(jobInstance.PreId)
		if err != nil {
			fmt.Println("111")
			log.Println(err)
			return
		}
		nodoInfo := depenConfig.Nodes[jobInstance.Config.ID]
		nodoInfo.Status = status
		depenConfig.Nodes[jobInstance.Config.ID] = nodoInfo
		fmt.Println(depenConfig)
		findAndExcute(jobInstance.PreId, &depenConfig)
		_, err2 := saveDepenConfig(jobInstance.PreId, depenConfig)
		if err2 != nil {
			log.Println(err2)
		}
		fmt.Printf("lockA Unlock %v \n", index)
		locks[index].Unlock()
	}
}

func findAndExcute(jobInstanceID string, depenConfig *proto.Dependency) {
	tasks := getAvaiTask(*depenConfig)
	//如果没有获取到可以执行的任务
	if len(tasks) == 0 {
		allEnd := true
		isFail := false
		//判断是否所有子任务都已经执行完毕
		for _, task := range depenConfig.Nodes {
			if task.Status != constants.TASK_COMPLETED {
				allEnd = false
			}
			if task.Status == constants.TASK_FAIL {
				isFail = true
				break
			}
		}
		if isFail {
			instance, _ := getInstanceConfig(jobInstanceID)
			event := eventcenter.NewEvent().WithBody(instance).WithHeader("status", constants.TASK_FAIL)
			eventCenter.Publish("job complete", event)
			fmt.Printf("依赖任务---%s---失败\n", jobInstanceID)
		}
		if allEnd {
			instance, _ := getInstanceConfig(jobInstanceID)
			event := eventcenter.NewEvent().WithBody(instance).WithHeader("status", constants.TASK_COMPLETED)
			eventCenter.Publish("job complete", event)
			fmt.Printf("依赖任务---%s---已完成\n", jobInstanceID)
		}
	}
	for _, task := range tasks {
		//任务执行时更新状态
		nodeInfo := depenConfig.Nodes[task]
		nodeInfo.Status = constants.TASK_RUNNING
		fmt.Println(task)
		config, err := configCenter.Get(task)
		fmt.Println(config)
		if err != nil {
			nodeInfo.Status = constants.TASK_UNEXECUTED
			log.Fatal(err)
			continue
		}
		taskConfig := proto.Config{}
		err = json.Unmarshal([]byte(config.Content), &taskConfig)
		if err != nil {
			nodeInfo.Status = constants.TASK_UNEXECUTED
			log.Println(err)
			continue
		}
		saveDepenConfig(jobInstanceID, *depenConfig)
		//go test(jobInstanceID, mainID, taskConfig.ID)
		instance := proto.JobInstance{
			Config:       taskConfig,
			ID:           strconv.FormatInt(sf.GenerateID(), 10),
			IsMaster:     true,
			ExecuteCount: 1,
			ExecuteTime:  time.Now().Unix(),
			StartTime:    time.Now().Unix(),
			Type:         constants.DEPENDENCE_SUB_JOB,
			PreId:        jobInstanceID,
			Status:       constants.TASK_RUNNING,
		}
		saveInstanceConfig(instance)
		nodeInfo.InstanceID = instance.ID
		depenConfig.Nodes[task] = nodeInfo
		//test(&instance)
		go ExecuteSingleTask(&instance)
	}
}

func saveInstanceConfig(instance proto.JobInstance) {
	instanceConfig, _ := json.Marshal(instance)
	configCenter.Save(configcenter.Config{
		ID:      instance.ID,
		Content: string(instanceConfig),
	})
}

func getInstanceConfig(id string) (proto.JobInstance, error) {
	//todo 此处应该加上group参数
	config, _ := configCenter.Get(id)
	instanceConfig := proto.JobInstance{}
	err := json.Unmarshal([]byte(config.Content), &instanceConfig)
	if err != nil {
		fmt.Println("id:" + id)
		fmt.Println(err)
		return proto.JobInstance{}, err
	}
	return instanceConfig, nil
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
