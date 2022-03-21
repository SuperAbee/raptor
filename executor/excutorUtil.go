package executor

import (
	"bytes"
	"io"
	"io/ioutil"
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
