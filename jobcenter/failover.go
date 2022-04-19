package jobcenter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"raptor/dependence"
	"raptor/eventcenter"
	"raptor/proto"
	"raptor/servicecenter"
)

func onJobTimeout(event *eventcenter.Event) {
	instance := event.Body.(proto.JobInstance)
	replace(instance)
}

func replace(instance proto.JobInstance) {
	jobCenter := New()
	log.Println(instance.ID + " timeout check master state")

	deleyExecutor, err := dependence.NewDeleyExecutor()
	if err != nil {
		log.Println(err.Error())
	}

	//依次检查前置节点健康状态
	cc := jobCenter.ConfigCenter
	content, _ := cc.Get(instance.Config.ID)
	var runningJob RunningJob
	json.Unmarshal([]byte(content.Content), &runningJob)

	var remainNodes []Node
	for i, node := range runningJob.Hosts {
		if node.Ip == jobCenter.Ip && node.Port == jobCenter.Port {
			remainNodes = runningJob.Hosts[i:]
			break
		}
		if err := ping(node); err == nil {
			//前置节点健康，由前置节点接替主节点
			return
		}
	}

	//前置节点失联，接替主节点
	log.Println("previous nodes connect failed, replace master")
	//补充不足的从节点
	fillUpNodes(remainNodes, 3, runningJob.Config.ID)
	runningJob.Hosts = remainNodes
	//作为主节点执行任务
	instance.IsMaster = true
	deleyExecutor.AddOrRun(instance)
}

func fillUpNodes(nodes []Node, aimLen int, jobID string) {
	//补充从节点
	if len(nodes) == aimLen {
		return
	}
	Scheduler, err := jobCenter.ServiceCenter.GetService("scheduler")
	if err != nil {
		panic(err)
	}
	selectedHosts := selectHosts(Scheduler)

	if aimLen > len(selectedHosts) {
		aimLen = len(selectedHosts)
	}

	for i := 0; len(nodes) < aimLen; i++ {
		if !contains(nodes, selectedHosts[i]) {
			nodes = append(nodes, Node{selectedHosts[i].Ip, selectedHosts[i].Port, false})
			//通知对应节点
			url := fmt.Sprintf("http://%s:%v%s?isMaster=%v&id=%s", selectedHosts[i].Ip, selectedHosts[i].Port, timingUrl, false, jobID)
			http.Get(url)
		}
	}
}

func contains(nodes []Node, host servicecenter.Instance) bool {
	for _, node := range nodes {
		if node.Ip == host.Ip {
			return true
		}
	}
	return false
}

func ping(host Node) error {
	url := fmt.Sprintf("http://%s:%v%s", host.Ip, host.Port, pingUrl)
	_, err := http.Get(url)
	return err
}
