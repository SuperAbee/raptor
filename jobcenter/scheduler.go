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
	pingUrl   = "/scheduler/ping"
	timingUrl = "/scheduler/timing"
	notifyUrl = "/scheduler/notify"
)

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

	j.RunningJobs.Store(jobID, runningJob)
	instance := generateJobInstance(runningJob.Config, isMaster)
	log.Printf("receieve job ID:%v name:%v isMaster:%v", jobID, runningJob.Config.Name, isMaster)

	//监听任务配置改变
	j.ConfigCenter.OnChange(jobID, onJobChange)

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

	log.Printf("generate new instance ID:%v jobID:%v", ID, config.ID)
	return instance
}

func onJobStart(event *eventcenter.Event) {
	jobCenter := New()
	instance := event.Body.(proto.JobInstance)

	//计算下一次任务实例
	//todo master重新上线的情况, 线程安全问题
	var newInstance proto.JobInstance
	if rj, ok := jobCenter.RunningJobs.Load(instance.Config.ID); ok {
		newInstance = generateJobInstance(rj.(RunningJob).Config, instance.IsMaster)
	}

	if instance.IsMaster {
		//检查从服务器状态并及时替换
		var hosts []Node
		if rj, ok := jobCenter.RunningJobs.Load(instance.Config.ID); ok {
			hosts = rj.(RunningJob).Hosts
		}

		for i := 0; i < len(hosts); i++ {
			if _, ok := jobCenter.Schedulers.Load(hosts[i]); !ok {
				hosts = append(hosts[:i], hosts[i+1:]...)
			}
		}

		fillUpNodes(hosts, 3, instance.Config.ID)

		//通知从节点任务状态
		notifyJob(instance, "running")
	}

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

	jobCenter.RunningJobs.Store(config.ID, runningJob)
}

func onJobFinished(event *eventcenter.Event) {
	state := event.Header["status"]
	instance := event.Body.(proto.JobInstance)

	//记录结果信息到nacos
	if instance.IsMaster {
		if state == "completed" {
			log.Println(instance.ID + " success")
		} else {
			log.Println(instance.ID + " success")
		}
	}

}

func notifyJob(instance proto.JobInstance, result string) {
	//通知从节点任务结果
	cc := jobCenter.ConfigCenter
	content, _ := cc.Get(instance.Config.ID)

	var runningJob RunningJob
	json.Unmarshal([]byte(content.Content), &runningJob)
	for _, host := range runningJob.Hosts[1:] {
		if !host.IsMaster {
			url := fmt.Sprintf("http://%s:%v%s?id=%s&state=%s", host.Ip, host.Port, notifyUrl, instance.ID, result)
			http.Get(url)
		}
	}
}

func getNextTime(cron string) int64 {
	nextTime := cronexpr.MustParse(cron).Next(time.Now())
	return nextTime.Unix()
}
