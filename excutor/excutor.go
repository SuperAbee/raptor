package excutor

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"raptor/proto"
	"raptor/servicecenter"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Excutor interface {
	ExcutorTask(config proto.Config) bool
}

func ExcutorTask(config proto.Config) bool {
	sc, _ := servicecenter.New(servicecenter.Nacos)

	service, _ := sc.GetService(config.TargetService)

	hosts := service.Hosts

	var healthyHosts []servicecenter.Instance

	for _, instance := range hosts {
		if instance.Healthy == true {
			healthyHosts = append(healthyHosts, instance)
		}
	}

	n := len(healthyHosts)

	//当前无健康服务
	if n < 1 {
		return false
	}

	//该任务不分片,直接执行
	if reflect.DeepEqual(config.ShardingStrategy, proto.ShardingStrategy{}) {
		//组装url
		var build strings.Builder
		build.WriteString(healthyHosts[0].Ip)
		build.WriteString(":")
		build.WriteString(strconv.FormatUint(healthyHosts[0].Port, 10))
		build.WriteString(config.Task.URI)
		url := build.String()

		if config.Task.Type == "GET" {
			Get(url)
		} else {
			Post(url, config.Task.Body, config.Task.Header, "application/json")
		}
	} else {

	}
	return true
}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string) string {

	// 超时时间：5秒
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.String()
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// header:		 POST请求头内容
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func Post(url string, body string, header map[string]string, contentType string) string {

	var jsonStr = []byte(body)
	//fmt.Println("jsonStr", jsonStr)
	//fmt.Println("new_str", bytes.NewBuffer(jsonStr))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	// req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", contentType)

	if header != nil && len(header) > 0 {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	result, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(result))

	return string(result)
}
