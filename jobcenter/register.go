package jobcenter

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"raptor/configcenter"
	"raptor/constants"
	"raptor/proto"
	"raptor/servicecenter"
	"strconv"
	"time"
)

func (j *JobCenter) SaveJob(config proto.Config) error {
	if config.ID == "" {
		config.ID = strconv.FormatInt(sf.GenerateID(), 10)
	}

	contentJson, _ := json.Marshal(config)
	_, err := j.ConfigCenter.Save(configcenter.Config{
		ID:      config.ID,
		Content: string(contentJson),
	})
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("save job ID:%v name:%v\n", config.ID, config.Name)

	j.ConfigCenter.Save(configcenter.Config{
		ID:      config.ID,
		Group:   constants.JOB_GROUP,
		Content: string(contentJson),
	})
	return nil
}

func (j *JobCenter) Register(config proto.Config) error {
	config.ID = strconv.FormatInt(sf.GenerateID(), 10)
	//选取节点
	Scheduler, err := j.ServiceCenter.GetService("scheduler")
	if err != nil {
		log.Println(err.Error())
		return err
	}
	selectedHosts := selectHosts(Scheduler)
	nodes := []Node{}
	for i, host := range selectedHosts {
		var isMaster bool
		if i == 0 {
			isMaster = true
		} else {
			isMaster = false
		}
		nodes = append(nodes, Node{
			Ip:       host.Ip,
			Port:     host.Port,
			IsMaster: isMaster,
		})
	}

	//任务注册到注册中心
	runningJob := RunningJob{
		Config: config,
		Hosts:  nodes,
	}

	contentJson, _ := json.Marshal(runningJob)
	_, err = j.ConfigCenter.Save(configcenter.Config{
		ID:      config.ID,
		Content: string(contentJson),
	})
	if err != nil {
		log.Println(err.Error())
		return err
	}
	log.Printf("register job ID:%v name:%v\n", config.ID, config.Name)

	//将任务分配给对应节点
	j.AssignJob(&runningJob)

	return nil
}

func (j *JobCenter) AssignJob(runningJob *RunningJob) {
	//通知对应节点负责任务
	for _, host := range runningJob.Hosts {
		url := fmt.Sprintf("http://%s:%v%s?isMaster=%v&id=%s", host.Ip, host.Port, timingUrl, host.IsMaster, runningJob.Config.ID)
		http.Get(url)
	}
}

func (j *JobCenter) Unregister(jobName string) (bool, error) {
	//删除不会存在(
	return false, nil
}

func (j *JobCenter) GetJobData(id string) (proto.Config, error) {
	data, err := j.ConfigCenter.Get(id)
	if err != nil {
		return proto.Config{}, err
	}

	var config proto.Config

	if err := json.Unmarshal([]byte(data.Content), &config); err != nil {
		return proto.Config{}, err
	}
	return config, nil
}

func selectHosts(service servicecenter.Service) []servicecenter.Instance {
	//根据不同策略选择不同的节点
	return randomNodes(service)
}

func randomNodes(service servicecenter.Service) []servicecenter.Instance {
	nodes := make([]servicecenter.Instance, 3)
	hosts := service.Hosts
	if len(hosts) <= 3 {
		return hosts
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
