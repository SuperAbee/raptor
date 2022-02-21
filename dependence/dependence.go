package dependence

import (
	"encoding/json"
	"log"
	"raptor/configcenter"
	"raptor/constants"
	"raptor/excutor"
	"raptor/proto"
	"sync"
)

var lock sync.Mutex
var configCenter, _ = configcenter.New("k8s")

func Execute(config proto.Config) (bool, error) {
	lock.Lock()
	findAndExcute(config.ID, &config.Dependencies)
	_, err := saveDepenConfig(config.ID, config.Dependencies)
	if err != nil {
		lock.Unlock()
		return false, err
	}
	lock.Unlock()
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

func afterSingleComplete(mainID string, ID string, status string) {
	lock.Lock()
	depenConfig, err := getDepenConfig(mainID)
	if err != nil {
		log.Fatal(err)
		lock.Unlock()
		return
	}
	depenConfig.Nodes[ID] = status
	findAndExcute(mainID, &depenConfig)
	saveDepenConfig(mainID, depenConfig)
	lock.Unlock()
}

func findAndExcute(mainID string, depenConfig *proto.Dependency) {
	tasks := getAvaiTask(*depenConfig)
	for _, task := range tasks {
		depenConfig.Nodes[task] = constants.TASK_RUNNING
		//updateConfig(mainID, task, constants.TASK_RUNNING)
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
			log.Fatal(err)
			continue
		}
		go excutor.ExcutorTask(taskConfig)
	}
}

func updateConfig(mainID string, ID string, status string) {
	config, err := configCenter.Get(mainID + "Dependence")
	if err != nil {
		return
	}
	depen := proto.Dependency{}
	json.Unmarshal([]byte(config.Content), depen)
	depen.Nodes[ID] = status
	depenJson, _ := json.Marshal(depen)
	config.Content = string(depenJson)
	configCenter.Save(config)
	//Temp.Dependencies.Nodes[ID] = status
}

func getDepenConfig(ID string) (proto.Dependency, error) {
	config, err := configCenter.Get(ID + "Dependence")
	if err != nil {
		return proto.Dependency{}, err
	}
	depen := proto.Dependency{}
	err = json.Unmarshal([]byte(config.Content), depen)
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
