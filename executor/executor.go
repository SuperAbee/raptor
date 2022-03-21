package executor

import (
	"fmt"
	"log"
	"raptor/proto"
	"raptor/servicecenter"
	"reflect"
	"strings"
)

var Executors map[string]Executor
var sc = servicecenter.New()

const HttpExecutorKey = "http_executor"

func init() {
	go getMsg()
	Executors = make(map[string]Executor)
	Executors[HttpExecutorKey] = &HttpExecutor{}
}

type Executor interface {
	Execute(jobInstance proto.JobInstance) error
}

type HttpExecutor struct {
}

var chMsg = make(chan ShardingRequest, 10)

// ShardingRequest 分片完成返回的http请求结果
type ShardingRequest struct {
	Id             string            // 任务id
	ShardingCount  int               // 分片总数
	ShardingItem   int               // 分片序号
	ShardingStatus int               // 分片执行结果
	RetryCount     int               // 失败重试总数
	TargetService  string            // 被调服务
	Type           string            // GET/POST
	Url            string            // 请求的url
	Body           string            // 任务体
	Header         map[string]string // 任务头
}

// TaskRequest 任务完成情况
type TaskRequest struct {
	ShardingNowCount int   // 分片完成数量
	ShardingStatus   []int // 分片执行情况
	RetryNowCount    int   // 重试次数
}

//暂存正在执行的任务信息，用于后续回调
var taskResults = make(map[string]TaskRequest)

//消费管道消息
func getMsg() {
	for {
		ch := <-chMsg
		log.Println("Id:", ch.Id, ",分片号：", ch.ShardingItem, ",结果:", ch.ShardingStatus)
		taskRequest, ok := taskResults[ch.Id] /*如果确定是真实的,则存在,否则不存在 */

		if ok {
			//分片请求成功
			if ch.ShardingStatus < 400 {
				nowCount := 1 + taskRequest.ShardingNowCount
				shardingStatus := taskRequest.ShardingStatus
				shardingStatus[ch.ShardingItem] = ch.ShardingStatus
				if nowCount == len(shardingStatus) {
					//回调wyf的函数，告知任务执行结果
					log.Println("Id:", ch.Id, "执行结果:", shardingStatus[0:])
					delete(taskResults, ch.Id)
				} else {
					taskResults[ch.Id] = TaskRequest{
						nowCount,
						shardingStatus,
						taskRequest.RetryNowCount,
					}
				}
			} else { //分片请求失败
				nowCount := taskRequest.ShardingNowCount
				shardingStatus := taskRequest.ShardingStatus
				shardingStatus[ch.ShardingItem] = ch.ShardingStatus
				nowRetry := taskRequest.RetryNowCount

				if nowRetry >= ch.RetryCount { //重试次数用完仍失败
					nowCount++
					if nowCount == len(shardingStatus) {
						//回调wyf的函数，告知任务执行结果
						log.Println("Id:", ch.Id, "执行结果:", shardingStatus[0:])
						delete(taskResults, ch.Id)
					} else {
						taskResults[ch.Id] = TaskRequest{
							nowCount,
							shardingStatus,
							nowRetry,
						}
					}
				} else { //执行分片失败重试
					nowRetry++

					taskResults[ch.Id] = TaskRequest{
						nowCount,
						shardingStatus,
						nowRetry,
					}

					healthyHosts := GetHealthyHostsByName(ch.TargetService)
					k := len(healthyHosts) - 1
					DoShardingRetry(ch, healthyHosts[k].Ip, healthyHosts[k].Port)
				}
			}
		} else {
			//分片请求成功
			if ch.ShardingStatus < 400 {
				nowCount := 1
				shardingStatus := make([]int, ch.ShardingCount)
				shardingStatus[ch.ShardingItem] = ch.ShardingStatus
				if nowCount == len(shardingStatus) {
					//回调wyf的函数，告知任务执行结果
					log.Println("Id:", ch.Id, "执行结果:", shardingStatus[0:])

				} else {
					taskResults[ch.Id] = TaskRequest{
						nowCount,
						shardingStatus,
						0,
					}
				}
			} else { //分片请求失败
				nowCount := 0
				shardingStatus := make([]int, ch.ShardingCount)
				shardingStatus[ch.ShardingItem] = ch.ShardingStatus
				nowRetry := 0

				if nowRetry >= ch.RetryCount { //重试次数用完仍失败
					nowCount++
					if nowCount == len(shardingStatus) {
						//回调wyf的函数，告知任务执行结果
						log.Println("Id:", ch.Id, "执行结果:", shardingStatus[0:])
					} else {
						taskResults[ch.Id] = TaskRequest{
							nowCount,
							shardingStatus,
							nowRetry,
						}
					}
				} else { //执行分片失败重试
					nowRetry++

					taskResults[ch.Id] = TaskRequest{
						nowCount,
						shardingStatus,
						nowRetry,
					}

					healthyHosts := GetHealthyHostsByName(ch.TargetService)
					k := len(healthyHosts) - 1
					DoShardingRetry(ch, healthyHosts[k].Ip, healthyHosts[k].Port)
				}
			}
		}
	}
}

func DoShardingRetry(request ShardingRequest, ip string, port uint64) {
	url := GetUrl(ip, port, request.Url, "")
	if request.Type == "GET" {
		go func() {
			status, _ := Get(url)

			request.ShardingStatus = status

			chMsg <- request
		}()
	} else {
		go func() {
			status, _ := Post(url, request.Body, request.Header, "application/json")

			request.ShardingStatus = status

			chMsg <- request
		}()
	}
}

// Execute 异步执行
func (h *HttpExecutor) Execute(jobInstance proto.JobInstance) error {
	config := jobInstance.Config
	//healthyHosts := [1]servicecenter.Instance{
	//	{
	//		"127.0.0.1",
	//		1234,
	//		true}}

	healthyHosts := GetHealthyHosts(config)

	healthyNum := len(healthyHosts)
	//当前无健康服务
	if healthyNum < 1 {
		return fmt.Errorf("no avaliable service")
	}

	//该任务不分片,直接执行
	if reflect.DeepEqual(config.ShardingStrategy, proto.ShardingStrategy{}) {

		url := GetUrl(healthyHosts[0].Ip, healthyHosts[0].Port, config.Task.URI, "")
		halfUrl := GetHalfUrl(config.Task.URI, "")

		if config.Task.Type == "GET" {
			go func() {
				status, _ := Get(url)
				shardingRequest := ShardingRequest{
					config.ID,
					1,
					0,
					status,
					jobInstance.ExecuteCount,
					config.TargetService,
					"GET",
					halfUrl,
					config.Task.Body,
					config.Task.Header,
				}
				chMsg <- shardingRequest
			}()
		} else {
			go func() {
				status, _ := Post(url, config.Task.Body, config.Task.Header, "application/json")
				shardingRequest := ShardingRequest{
					config.ID,
					1,
					0,
					status,
					jobInstance.ExecuteCount,
					config.TargetService,
					"POST",
					halfUrl,
					config.Task.Body,
					config.Task.Header,
				}
				chMsg <- shardingRequest
			}()
			//Post(url, config.Task.Body, config.Task.Header, "application/json")
		}
	} else {
		// 动态分片 || 静态初次分片 || 静态重新分片
		if config.ShardingStrategy.ShardingType == "dynamic" || config.ShardingResults == nil || (config.ShardingStrategy.ActuallyCount < config.ShardingStrategy.DefaultCount && config.ShardingStrategy.ActuallyCount < healthyNum) {
			strs := strings.Split(config.ShardingStrategy.ParameterRole, ",")
			//1个分片怎么做？
			n := len(strs)
			parameters := make([]string, n)
			//0=A,1=B 从index2开始是参数
			for i, parameter := range strs {
				parameters[i] = string([]rune(parameter)[2:])
				//fmt.Println(parameters[i])
			}
			defaultNum := config.ShardingStrategy.DefaultCount
			var instanceNum = 0
			//取默认执行实例数、健康实例数中 小者 作为本次执行实例数
			if defaultNum > healthyNum {
				instanceNum = healthyNum
			} else {
				instanceNum = defaultNum
			}
			/*
			*	需要一个提交更改config的接口
			*
			 */
			config.ShardingStrategy.ActuallyCount = instanceNum

			shardingResults := make([]proto.Sharding, n)
			//实例数多于分片数，则1分片1实例
			if instanceNum >= n {
				for i := 0; i < n; i++ {
					xx := i
					url := GetUrl(healthyHosts[i].Ip, healthyHosts[i].Port, config.Task.URI, parameters[i])
					halfUrl := GetHalfUrl(config.Task.URI, parameters[i])
					if config.Task.Type == "GET" {
						go func() {
							status, _ := Get(url)
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"GET",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Get(url)
					} else {
						go func() {
							status, _ := Post(url, config.Task.Body, config.Task.Header, "application/json")
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"POST",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[i] = proto.Sharding{
							ShardingItem: i,
							Parameter:    parameters[i],
							Ip:           healthyHosts[i].Ip,
							Port:         healthyHosts[i].Port,
						}
					}
				}
			} else { //分片数多于实例数，平均分片
				k := n / instanceNum
				r := n % instanceNum
				p := k * instanceNum
				for i := 0; i < p; i++ {
					no := i % instanceNum
					url := GetUrl(healthyHosts[no].Ip, healthyHosts[no].Port, config.Task.URI, parameters[i])
					halfUrl := GetHalfUrl(config.Task.URI, parameters[i])
					xx := i
					if config.Task.Type == "GET" {
						go func() {
							status, _ := Get(url)
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"GET",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Get(url)
					} else {
						go func() {
							status, _ := Post(url, config.Task.Body, config.Task.Header, "application/json")
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"POST",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[i] = proto.Sharding{
							ShardingItem: i,
							Parameter:    parameters[i],
							Ip:           healthyHosts[no].Ip,
							Port:         healthyHosts[no].Port,
						}
					}
				}
				for i := 0; i < r; i++ {
					xx := i
					url := GetUrl(healthyHosts[i].Ip, healthyHosts[i].Port, config.Task.URI, parameters[p+i])
					halfUrl := GetHalfUrl(config.Task.URI, parameters[p+i])
					if config.Task.Type == "GET" {
						go func() {
							status, _ := Get(url)
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"GET",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Get(url)
					} else {
						go func() {
							status, _ := Post(url, config.Task.Body, config.Task.Header, "application/json")
							shardingRequest := ShardingRequest{
								config.ID,
								config.ShardingStrategy.ShardingCount,
								xx,
								status,
								jobInstance.ExecuteCount,
								config.TargetService,
								"POST",
								halfUrl,
								config.Task.Body,
								config.Task.Header,
							}
							chMsg <- shardingRequest
						}()
						//Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[p+i] = proto.Sharding{
							ShardingItem: p + i,
							Parameter:    parameters[p+i],
							Ip:           healthyHosts[i].Ip,
							Port:         healthyHosts[i].Port,
						}
					}
				}
			}
			//提交静态分片结果config
			if config.ShardingStrategy.ShardingType == "static" {
				config.ShardingResults = shardingResults
			}
		} else if config.ShardingStrategy.ShardingType == "static" {
			//已有静态分片结果，且不需要重新分片
			n := len(config.ShardingResults)
			for i := 0; i < n; i++ {
				xx := i
				url := GetUrl(config.ShardingResults[i].Ip, config.ShardingResults[i].Port, config.Task.URI, config.ShardingResults[i].Parameter)
				halfUrl := GetHalfUrl(config.Task.URI, config.ShardingResults[i].Parameter)
				if config.Task.Type == "GET" {
					go func() {
						status, _ := Get(url)
						shardingRequest := ShardingRequest{
							config.ID,
							config.ShardingStrategy.ShardingCount,
							xx,
							status,
							jobInstance.ExecuteCount,
							config.TargetService,
							"GET",
							halfUrl,
							config.Task.Body,
							config.Task.Header,
						}
						chMsg <- shardingRequest
					}()
					//Get(url)
				} else {
					go func() {
						status, _ := Post(url, config.Task.Body, config.Task.Header, "application/json")
						shardingRequest := ShardingRequest{
							config.ID,
							config.ShardingStrategy.ShardingCount,
							xx,
							status,
							jobInstance.ExecuteCount,
							config.TargetService,
							"POST",
							halfUrl,
							config.Task.Body,
							config.Task.Header,
						}
						chMsg <- shardingRequest
					}()
					//Post(url, config.Task.Body, config.Task.Header, "application/json")
				}
			}
		}
	}
	return nil
}

// SynExecute 同步执行(需要更新代码)
func (h *HttpExecutor) SynExecute(config proto.Config) error {

	healthyHosts := GetHealthyHosts(config)
	healthyNum := len(healthyHosts)
	//当前无健康服务
	if healthyNum < 1 {
		return fmt.Errorf("no avaliable service")
	}

	//该任务不分片,直接执行
	if reflect.DeepEqual(config.ShardingStrategy, proto.ShardingStrategy{}) {

		url := GetUrl(healthyHosts[0].Ip, healthyHosts[0].Port, config.Task.URI, "")

		if config.Task.Type == "GET" {
			Get(url)
		} else {
			Post(url, config.Task.Body, config.Task.Header, "application/json")
		}
	} else {
		// 动态分片 || 静态初次分片 || 静态重新分片
		if config.ShardingStrategy.ShardingType == "dynamic" || config.ShardingResults == nil || (config.ShardingStrategy.ActuallyCount < config.ShardingStrategy.DefaultCount && config.ShardingStrategy.ActuallyCount < healthyNum) {
			strs := strings.Split(config.ShardingStrategy.ParameterRole, ",")
			//1个分片怎么做？
			n := len(strs)
			parameters := make([]string, n)
			//0=A,1=B 从index2开始是参数
			for i, parameter := range strs {
				parameters[i] = string([]rune(parameter)[2:])
				//fmt.Println(parameters[i])
			}
			defaultNum := config.ShardingStrategy.DefaultCount
			var instanceNum = 0
			//取默认执行实例数、健康实例数中 小者 作为本次执行实例数
			if defaultNum > healthyNum {
				instanceNum = healthyNum
			} else {
				instanceNum = defaultNum
			}
			/*
			*	需要一个提交更改config的接口
			*
			 */
			config.ShardingStrategy.ActuallyCount = instanceNum

			shardingResults := make([]proto.Sharding, n)
			//实例数多于分片数，则1分片1实例
			if instanceNum >= n {
				for i := 0; i < n; i++ {
					url := GetUrl(healthyHosts[i].Ip, healthyHosts[i].Port, config.Task.URI, parameters[i])
					if config.Task.Type == "GET" {
						Get(url)
					} else {
						Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[i] = proto.Sharding{
							ShardingItem: i,
							Parameter:    parameters[i],
							Ip:           healthyHosts[i].Ip,
							Port:         healthyHosts[i].Port,
						}
					}
				}
			} else { //分片数多于实例数，平均分片
				k := n / instanceNum
				r := n % instanceNum
				p := k * instanceNum
				for i := 0; i < p; i++ {
					no := i % instanceNum
					url := GetUrl(healthyHosts[no].Ip, healthyHosts[no].Port, config.Task.URI, parameters[i])
					if config.Task.Type == "GET" {
						Get(url)
					} else {
						Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[i] = proto.Sharding{
							ShardingItem: i,
							Parameter:    parameters[i],
							Ip:           healthyHosts[no].Ip,
							Port:         healthyHosts[no].Port,
						}
					}
				}
				for i := 0; i < r; i++ {
					url := GetUrl(healthyHosts[i].Ip, healthyHosts[i].Port, config.Task.URI, parameters[p+i])
					if config.Task.Type == "GET" {
						Get(url)
					} else {
						Post(url, config.Task.Body, config.Task.Header, "application/json")
					}
					if config.ShardingStrategy.ShardingType == "static" {
						shardingResults[p+i] = proto.Sharding{
							ShardingItem: p + i,
							Parameter:    parameters[p+i],
							Ip:           healthyHosts[i].Ip,
							Port:         healthyHosts[i].Port,
						}
					}
				}
			}
			//提交静态分片结果config
			if config.ShardingStrategy.ShardingType == "static" {
				config.ShardingResults = shardingResults
			}
		} else if config.ShardingStrategy.ShardingType == "static" {
			//已有静态分片结果，且不需要重新分片
			n := len(config.ShardingResults)
			for i := 0; i < n; i++ {
				url := GetUrl(config.ShardingResults[i].Ip, config.ShardingResults[i].Port, config.Task.URI, config.ShardingResults[i].Parameter)
				if config.Task.Type == "GET" {
					Get(url)
				} else {
					Post(url, config.Task.Body, config.Task.Header, "application/json")
				}
			}
		}
	}
	return nil
}
