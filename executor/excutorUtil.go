package executor

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"raptor/proto"
	"raptor/servicecenter"
	"strconv"
	"strings"
	"time"
)

// GetHealthyHosts 获取当前访问服务的健康实例表
func GetHealthyHosts(config proto.Config) []servicecenter.Instance {
	service, _ := sc.GetService(config.TargetService)
	hosts := service.Hosts
	var healthyHosts []servicecenter.Instance
	for _, instance := range hosts {
		if instance.Healthy == true {
			healthyHosts = append(healthyHosts, instance)
		}
	}
	return healthyHosts
}

// GetRandomHosts 获取随机排列后的健康实例表
func GetRandomHosts(list []servicecenter.Instance) []servicecenter.Instance {

	//生成[0,size)随机序列randoms
	size := len(list)
	rand.Seed(time.Now().UnixNano())
	randoms := rand.Perm(size)

	res := make([]servicecenter.Instance, size)
	for i := 0; i < size; i++ {
		res[i] = list[randoms[i]]
	}

	return res
}

func GetRandomHostsByShuffle(list []servicecenter.Instance) []servicecenter.Instance {
	for i := len(list) - 1; i > 0; i-- {
		rand.Seed(time.Now().UnixNano())
		k := rand.Intn(i + 1)
		e := list[k]
		list[k] = list[i]
		list[i] = e
	}
	return list
}

// GetHealthyHostsByName 获取当前访问服务的健康实例表
func GetHealthyHostsByName(targetService string) []servicecenter.Instance {

	service, _ := sc.GetService(targetService)

	hosts := service.Hosts

	var healthyHosts []servicecenter.Instance

	for _, instance := range hosts {
		if instance.Healthy == true {
			healthyHosts = append(healthyHosts, instance)
		}
	}

	return healthyHosts
}

func GetUrl(ip string, port uint64, uri string, parameter string) string {
	//组装url
	var build strings.Builder
	build.WriteString("http://")
	build.WriteString(ip)
	build.WriteString(":")
	build.WriteString(strconv.FormatUint(port, 10))
	build.WriteString(uri)
	if parameter != "" {
		build.WriteString("/" + parameter)
	}
	return build.String()
}

func GetHalfUrl(uri string, parameter string) string {
	//组装url
	var build strings.Builder
	build.WriteString(uri)
	if parameter != "" {
		build.WriteString("/" + parameter)
	}
	return build.String()
}

// Get 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string) (int, string) {

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

	status, _ := strconv.Atoi(resp.Status[0:3])

	return status, result.String()
}

// Post 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// header:		 POST请求头内容
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func Post(url string, body string, header map[string]string, contentType string) (int, string) {

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

	status, _ := strconv.Atoi(resp.Status[0:3])

	return status, string(result)
}
