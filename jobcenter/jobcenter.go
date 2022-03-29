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
	"time"
)

type JobRegistry interface {
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
	RunningJobs   map[string]RunningJob
	Ip            string
	Port          uint64
	State         string
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
	sf, _ = uuid.NewSnowFlakeUUID((time.Now().Unix() % 1024) + 1)

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
	rj := make(map[string]RunningJob)

	jobCenter = JobCenter{
		ServiceCenter: sc,
		ConfigCenter:  cc,
		EventCenter:   es,
		RunningJobs:   rj,
		Ip:            ip,
		Port:          1234,
		State:         "running",
	}

	//监听任务状态
	jobCenter.EventCenter.Subscribe("jobStart", onJobStart)
	jobCenter.ConfigCenter.OnChange("jobChange", onJobChange)
	jobCenter.EventCenter.Subscribe("jobFinished", onJobFinished)
	jobCenter.EventCenter.Subscribe("jobTimeout", onJobTimeout)

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
