package jobcenter

import (
	"encoding/json"
	"log"
	"raptor/configcenter"
	"raptor/proto"
)

func (j *JobCenter) Register(config proto.Config) error {
	//任务注册到注册中心
	configJson, _ := json.Marshal(config)
	_, err := j.configCenter.Save(configcenter.Config{
		ID:      config.ID,
		Content: string(configJson),
	})
	if err != nil {
		return err
	}
	log.Print("register Job:", config.Name)

	//将任务分配给对应节点
	j.AssignJob(&config)

	return nil
}

func (j *JobCenter) Unregister(jobName string) (bool, error) {
	//删除不会存在(
	return false, nil
}

func (j *JobCenter) GetJobData(id string) (proto.Config, error) {
	data, err := j.configCenter.Get(id)
	if err != nil {
		return proto.Config{}, err
	}

	var config proto.Config

	if err := json.Unmarshal([]byte(data.Content), &config); err != nil {
		return proto.Config{}, err
	}
	return config, nil
}
