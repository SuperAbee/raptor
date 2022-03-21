package jobcenter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"raptor/dependence"
	"raptor/proto"
)

func onJobFailed(instance proto.JobInstance) {
	log.Println(instance.ID + " failed, start failover")
	deleyExecutor, err := dependence.NewDeleyExecutor()
	if err != nil {
		log.Println(err.Error())
	}

	if instance.IsMaster {
		//主节点重试并通知从节点任务结果
		log.Println("retry ", instance.ID)
		instance.ExecuteCount++
		deleyExecutor.AddOrRun(instance)
		notifyJobResult(instance, "failed")
	} else {
		//从节点检查主节点健康状态
		cc := jobCenter.ConfigCenter
		content, _ := cc.Get(instance.Config.ID)
		var runningJob RunningJob
		json.Unmarshal([]byte(content.Content), &runningJob)
		var host Node
		for _, node := range runningJob.Hosts {
			if node.IsMaster {
				host = node
				break
			}
		}
		url := fmt.Sprintf("%s:%v%s", host.Ip, host.Port, pingUrl)
		if _, err := http.Get(url); err != nil {
			//主节点失联, 进入选举流程
			log.Println("master connect failed, start election")
			if election(&runningJob) {
				//如果选举成功则继续主节点工作
				log.Println("retry ", instance.ID)
				instance.ExecuteCount++
				deleyExecutor.AddOrRun(instance)
			}
		}
	}
}

func election(runningJob *RunningJob) bool {
	//进行选举
	return true
}
