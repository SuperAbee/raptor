package dependence

import (
	"fmt"
	"raptor/constants"
	"raptor/eventcenter"
	"raptor/proto"
	"reflect"
	"time"
)

type DeleyExecutor struct {
	tw            *TimeWheel
	skipList      *SkipList
	slaveInstance map[string]proto.JobInstance
}

func NewDeleyExecutor() (*DeleyExecutor, error) {
	timeWheel, err := NewTimeWheel(1*time.Second, 30)
	if err != nil {
		return nil, err
	}
	timeWheel.Start()
	return &DeleyExecutor{
		tw:       timeWheel,
		skipList: NewSkipList(),
		slaveInstance: make(map[string]proto.JobInstance),
	}, nil
}

func (dE *DeleyExecutor) Init() {
	dE.tw.AddCron(15*time.Second, func() {
		list := dE.skipList.getLessThan(time.Now().Unix() + 15)
		for k, _ := range list {
			fmt.Printf("将任务 %s 加入到时间轮中", k.ID)
			dE.AddOrRun(*k)
		}
	})
}

func (dE *DeleyExecutor) AddOrRun(jobInstance proto.JobInstance) (bool, error) {
	if !jobInstance.IsMaster {
		dE.slaveInstance[jobInstance.ID] = jobInstance
		jobInstance.ExecuteTime += 5
	}
	duration := jobInstance.ExecuteTime - time.Now().Unix()
	if duration <= 0 && !jobInstance.IsMaster {
		duration = 5
	}
	if duration <= 0 {
		dE.executeTask(&jobInstance)
		return true, nil
	} else if duration <= 30 {
		fmt.Printf("将任务 %s 加入时间轮\n", jobInstance.ID)
		if !jobInstance.IsMaster {
			dE.slaveInstance[jobInstance.ID] = jobInstance
			jobInstance.ExecuteTime += 5
		}
		dE.tw.Add(time.Duration(duration)*time.Second, func() {
			dE.executeTask(&jobInstance)
		})
		return true, nil
	} else {
		fmt.Printf("将任务 %s 进行存储\n", jobInstance.ID)
		_, err := dE.skipList.Insert(&jobInstance, jobInstance.ExecuteTime)
		if err != nil {
			return false, fmt.Errorf("task %s insert error", jobInstance.Config.ID)
		}
		return true, nil
	}
}

func (dE *DeleyExecutor) ChangeSalveInstance(instanceID string) {
	delete(dE.slaveInstance, instanceID)
}

func (dE *DeleyExecutor) executeTask(jobInstance *proto.JobInstance) {
	if !jobInstance.IsMaster {
		_, ok := dE.slaveInstance[jobInstance.ID]
		if !ok {
			event := eventcenter.NewEvent().WithBody(jobInstance)
			eventCenter.Publish("job timeout", event)
			return
		}
	}
	switch reflect.DeepEqual(jobInstance.Config.Dependencies, proto.Dependency{}) {
	case true:
		jobInstance.Type = constants.SINGLE_JOB
		go ExecuteSingleTask(jobInstance)
	case false:
		jobInstance.Type = constants.DEPENDENCE_JOB
		go ExecuteDependenceTask(jobInstance)
	}

}
