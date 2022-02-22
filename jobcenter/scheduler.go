package jobcenter

import (
	"log"
	"math/rand"
	"net/http"
	"raptor/proto"
	"raptor/servicecenter"
	"strconv"
	"time"
)

const (
	timingUrl = "/scheduler/timing"
)

func (j *JobCenter) AssignJob(config *proto.Config) error {
	Scheduler, err := j.serviceCenter.GetService("scheduler")
	if err != nil {
		log.Println(err.Error())
		return err
	}

	//随机选取并通知对应节点负责任务
	nodes := randomNodes(Scheduler)
	for i := 0; i < len(nodes); i++ {
		isMaster := 0
		if i == 0 {
			isMaster = 1
		}
		url := timingUrl + "?isMaster=" + strconv.FormatInt(int64(isMaster), 10) + "&id=" + config.ID
		http.Get(url)
	}

	return nil
}

func (j *JobCenter) TimingJob(isMaster, jobID string) error {
	log.Println("Start Timing Job")

	return nil
}

func randomNodes(service servicecenter.Service) map[int]servicecenter.Instance {
	nodes := make(map[int]servicecenter.Instance)
	hosts := service.Hosts
	if len(hosts) <= 3 {
		for i, host := range hosts {
			nodes[i] = host
		}
	}

	for i, num := range generateRandomNumber(0, len(hosts), 3) {
		nodes[i] = hosts[num]
	}

	return nodes
}

func generateRandomNumber(start int, end int, count int) []int {
	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn((end - start)) + start
		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}
