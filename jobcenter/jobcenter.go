package jobcenter

import (
	"log"
	"net"
	"raptor/configcenter"
	"raptor/eventcenter"
	"raptor/proto"
	"raptor/servicecenter"
	"raptor/uuid"
	"strings"
	"sync"
	"time"
)

type JobRegistry interface {
	AddJob(config proto.Config)

	Register(config proto.Config) error
	Unregister(jobName string) error
	GetJobData(jobName string) (proto.Config, error)
}

type Scheduler interface {
	Assign(config proto.Config) error
}

type JobCenter struct {
	ServiceCenter servicecenter.ServiceCenter
	ConfigCenter  configcenter.ConfigCenter
	EventCenter   *eventcenter.EventCenter
	RunningJobs   sync.Map
	Schedulers    sync.Map

	Ip    string
	Port  uint64
	State string
}

type RunningJob struct {
	Config proto.Config
	Hosts  []Node
}

type Node struct {
	Ip       string
	Port     uint64
	IsMaster bool
}

var jobCenter JobCenter

var sf *uuid.SnowFlakeUUID

func New() *JobCenter {
	return &jobCenter
}

func init() {
	log.Println("jobcenter init")
	var err error

	sf, err = uuid.NewSnowFlakeUUID((time.Now().Unix() % 1024) + 1)
	if err != nil {
		log.Fatalln(err.Error())
	}

	//获取注册中心和配置中心
	sc := servicecenter.New()
	cc := configcenter.New()

	//将本服务注册到注册中心
	ip, err := GetOutBoundIP()
	if err != nil {
		log.Println("can't get local IP")
		panic(err)
	}
	sc.Register(servicecenter.RegisterParam{
		Ip:          ip,
		Port:        1234,
		ServiceName: "scheduler",
	})

	//事件中心
	es := eventcenter.New()

	//当前服务运行的任务
	var rj sync.Map

	jobCenter = JobCenter{
		ServiceCenter: sc,
		ConfigCenter:  cc,
		EventCenter:   es,
		RunningJobs:   rj,
		Ip:            ip,
		Port:          1234,
		State:         "running",
	}

	//定时同步调度器状态
	go jobCenter.syncSchedulerList()

	//监听任务状态
	jobCenter.EventCenter.Subscribe("job start", onJobStart)
	jobCenter.EventCenter.Subscribe("job finished", onJobFinished)
	jobCenter.EventCenter.Subscribe("job timeout", onJobTimeout)

}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func (j *JobCenter) syncSchedulerList() {
	ticker := time.NewTicker(time.Second * 1)

	syncOnce := func() {
		service, err := j.ServiceCenter.GetService("scheduler")
		if err != nil {
			log.Fatalln(err.Error())
		}

		var newList sync.Map
		for _, instance := range service.Hosts {
			if instance.Healthy {
				newList.Store(Node{instance.Ip, instance.Port, false}, false)
				newList.Store(Node{instance.Ip, instance.Port, true}, false)
			}
		}
		j.Schedulers = newList

	}

	syncOnce()

	for range ticker.C {
		syncOnce()
	}

}
