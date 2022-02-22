package jobcenter

import (
	"log"
	"net"
	"raptor/configcenter"
	"raptor/proto"
	"raptor/servicecenter"
	"strings"
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
	serviceCenter servicecenter.ServiceCenter
	configCenter  configcenter.ConfigCenter
}

var jobCenter JobCenter

func New() *JobCenter {
	return &jobCenter
}

func init() {
	//获取注册中心和配置中心
	sc, err := servicecenter.New(servicecenter.Nacos)
	if err != nil {
		log.Println(err.Error())
	}

	cc, err := configcenter.New(configcenter.Nacos)
	if err != nil {
		log.Println(err.Error())
	}

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

	jobCenter = JobCenter{sc, cc}
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
