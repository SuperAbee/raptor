package jobcenter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"raptor/configcenter"
	"raptor/dependence"
	"raptor/eventcenter"
	"raptor/proto"
	"strconv"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	pingUrl     = "/scheduler/ping"
	timingUrl   = "/scheduler/timing"
	jobStateUrl = "/scheduler/state"
)

func (j *JobCenter) AssignJob(runningJob *RunningJob) error {
	//通知对应节点负责任务
	for _, host := range runningJob.Hosts {
		url := fmt.Sprintf("http://%s:%v%s?isMaster=%v&id=%s", host.Ip, host.Port, timingUrl, host.IsMaster, runningJob.Config.ID)
		http.Get(url)
	}

	return nil
}

func (j *JobCenter) TimingJob(isMaster bool, jobID string) error {
	content, err := j.ConfigCenter.Get(jobID)
	if err != nil {
		return err
	}

	//记录本机负责的任务
	var runningJob RunningJob
	err = json.Unmarshal([]byte(content.Content), &runningJob)
	if err != nil {
		return err
	}
	j.RunningJobs[jobID] = runningJob
	instance := generateJobInstance(runningJob.Config, isMaster)

	//执行任务
	deleyExecutor, err := dependence.NewDeleyExecutor()
	if err != nil {
		log.Println(err.Error())
	}
	deleyExecutor.AddOrRun(instance)

	return nil
}

func generateJobInstance(config proto.Config, isMaster bool) proto.JobInstance {
	//解析cron表达式
	time := getNextTime(config.Cron)
	//随机生成ID
	ID := strconv.FormatInt(sf.GenerateID(), 10)

	instance := proto.JobInstance{
		Config:       config,
		ExecuteTime:  int64(time),
		ID:           ID,
		IsMaster:     isMaster,
		ExecuteCount: 1,
	}
	return instance
}

func onJobStart(event *eventcenter.Event) {
	jobCenter := New()
	instance := event.Body.(proto.JobInstance)
	//计算下一次任务实例
	//todo master重新上线的情况, 线程安全问题
	newInstance := generateJobInstance(jobCenter.RunningJobs[instance.ID].Config, instance.IsMaster)

	//执行任务
	deleyExecutor, err := dependence.NewDeleyExecutor()
	if err != nil {
		log.Println(err.Error())
	}
	deleyExecutor.AddOrRun(newInstance)
}

func onJobChange(config configcenter.Config) {
	//同步任务信息到本地
	jobCenter := New()
	var runningJob RunningJob
	json.Unmarshal([]byte(config.Content), &runningJob)
	jobCenter.RunningJobs[config.ID] = runningJob
}

func onJobFinished(event *eventcenter.Event) {
	isSuccess := event.Header["isSuccess"]
	instance := event.Body.(proto.JobInstance)

	//记录结果信息到nacos
	if instance.IsMaster {
		if isSuccess == "true" {
			log.Println(instance.ID + " success")
		} else {
			log.Println(instance.ID + " success")
		}
	}

}

func notifyJobResult(instance proto.JobInstance, result string) {
	//通知从节点任务结果
	cc := jobCenter.ConfigCenter
	content, _ := cc.Get(instance.Config.ID)

	var runningJob RunningJob
	json.Unmarshal([]byte(content.Content), &runningJob)
	for _, host := range runningJob.Hosts {
		if !host.IsMaster {
			url := fmt.Sprintf("http://%s:%v%s?id=%s&state=%s", host.Ip, host.Port, jobStateUrl, instance.ID, result)
			http.Get(url)
		}
	}
}

func getNextTime(cron string) int64 {
	nextTime := cronexpr.MustParse(cron).Next(time.Now())
	return nextTime.Unix()
}
