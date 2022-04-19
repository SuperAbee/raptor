package router

import (
	"encoding/json"
	"log"
	"net/http"
	"raptor/configcenter"
	"raptor/constants"
	"raptor/dependence"
	"raptor/jobcenter"
	"raptor/proto"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Route(router *gin.Engine) {
	router.POST("/register", registerWorkflow)
	router.GET("/jobData", getJobData)

	scheduling(router.Group("/scheduler"))
	openInterface(router.Group("/raptor"))
}

func scheduling(router *gin.RouterGroup) {
	jobcenter.New()
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})
	router.GET("/timing", timingJob)
	//通知任务开始
	router.GET("/notify", updateJobState)
}

func openInterface(router *gin.RouterGroup) {
	router.POST("/user/login", login)

	//状态查询
	//status := router.Group("/status")
	//status.GET("/system")
	//status.GET("/machine")

	//任务操作
	job := router.Group("/job")
	job.POST("/add", saveJob)
	job.POST("/save", saveJob)
	job.GET("/search", searchJob)

	//任务依赖
	dep := router.Group("/dependency")
	dep.POST("/add", registerWorkflow)
	dep.POST("/save", saveWorkflow)
	dep.GET("/search", searchWorkflow)
	dep.GET("/instance", getJobInstance)

	//这个不太清楚
	//dep.GET("/subinstance")

	//重试待实现
	//dep.POST("/retry")
	//dep.POST("/subinstance/retry")
}

func login(c *gin.Context) {
	//登录信息怎么存
}

func registerWorkflow(c *gin.Context) {
	registry := jobcenter.New()

	var body proto.Config
	c.ShouldBind(&body)
	if err := registry.Register(body); err != nil {
		c.String(http.StatusOK, "register failed: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, struct {
		code int
		msg  string
	}{10000, "success"})
}

func saveWorkflow(c *gin.Context) {
	var body proto.Config
	c.ShouldBind(&body)
	cc := configcenter.New()

	content, err := cc.Get(body.ID)
	if err != nil {
		log.Println(err.Error())
	}

	//拿到运行任务
	var runningJob jobcenter.RunningJob
	err = json.Unmarshal([]byte(content.Content), &runningJob)
	if err != nil {
		log.Println(err.Error())
	}

	//改一下任务配置
	runningJob.Config = body

	//存回去
	contentJson, _ := json.Marshal(runningJob)
	_, err = cc.Save(configcenter.Config{
		ID:      body.ID,
		Content: string(contentJson),
	})
	if err != nil {
		log.Println(err.Error())
	}

	c.JSON(http.StatusOK, struct {
		code int
		msg  string
	}{10000, "success"})
}

func searchWorkflow(c *gin.Context) {
	//fixme 存储和查询都有点问题

	// jobId := c.Query("id")
	key := c.Query("key")

	// pageIndex := c.Query("pageIndex")
	// pageSize := c.Query("pageSize")
	// pageTotal := c.Query("pageTotal")

	var jobList []proto.Config

	cc := configcenter.New()
	results, err := cc.GetByKV(map[string]configcenter.Search{
		"Config.Name": {Keyword: key, Exact: false},
	}, "raptor")
	if err != nil {
		log.Println(err.Error())
	}
	for _, v := range results {
		var runningJob jobcenter.RunningJob
		err = json.Unmarshal([]byte(v.Content), &runningJob)
		if err != nil {
			log.Println(err.Error())
		}

		jobList = append(jobList, runningJob.Config)
	}

	c.JSON(http.StatusOK, struct {
		code int
		msg  string
		data interface{}
	}{10000, "success", jobList})
}

func getJobData(c *gin.Context) {
	jobRegister := jobcenter.New()

	id := c.Query("id")
	jobData, err := jobRegister.GetJobData(id)
	if err != nil {
		log.Print(err.Error())
		c.String(http.StatusOK, "can't find job")
		return
	}

	c.JSON(http.StatusOK, jobData)
}

func timingJob(c *gin.Context) {
	scheduler := jobcenter.New()

	isMaster := c.Query("isMaster")
	jobID := c.Query("id")

	masterParam, _ := strconv.ParseBool(isMaster)

	if err := scheduler.TimingJob(masterParam, jobID); err != nil {
		log.Println(err.Error())
	}
}

func updateJobState(c *gin.Context) {
	jobInstanceID := c.Query("id")
	state := c.Query("state")

	//执行器提供更新任务结果的接口
	deleyExecutor, err := dependence.NewDeleyExecutor()
	if err != nil {
		log.Println(err.Error())
	}
	deleyExecutor.ChangeSalveInstance(jobInstanceID)
	log.Printf("updateJob: id = %v, state = %v\n", jobInstanceID, state)
}

func saveJob(c *gin.Context) {
	registry := jobcenter.New()
	var body proto.Config
	c.ShouldBind(&body)
	if err := registry.SaveJob(body); err != nil {
		c.JSON(http.StatusOK, struct {
			code int
			msg  string
		}{10000, "save job failed: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, struct {
		code int
		msg  string
	}{10000, "success"})
}

func searchJob(c *gin.Context) {
	//fixme 不太清楚怎么查询的

	// jobId := c.Query("id")
	key := c.Query("key")

	// pageIndex := c.Query("pageIndex")
	// pageSize := c.Query("pageSize")
	// pageTotal := c.Query("pageTotal")

	var jobList []proto.Config

	cc := configcenter.New()
	results, err := cc.GetByKV(map[string]configcenter.Search{
		"Name": {Keyword: key, Exact: false},
	}, constants.JOB_GROUP)
	if err != nil {
		log.Println(err.Error())
	}
	for _, v := range results {
		var config proto.Config
		err = json.Unmarshal([]byte(v.Content), &config)
		if err != nil {
			log.Println(err.Error())
		}
		jobList = append(jobList, config)
	}

	c.JSON(http.StatusOK, struct {
		code int
		msg  string
		data interface{}
	}{10000, "success", jobList})
}

func getJobInstance(c *gin.Context) {
	//fixme 假设用jobID查
	jobId := c.Query("id")
	cc := configcenter.New()

	results, err := cc.GetByKV(map[string]configcenter.Search{
		"Config.ID": {Keyword: jobId, Exact: true},
	}, constants.JOB_INSTANCE_GROUP)
	if err != nil {
		log.Println(err.Error())
	}

	var instanceList []proto.JobInstance

	for _, v := range results {
		var instance proto.JobInstance
		err = json.Unmarshal([]byte(v.Content), &instance)
		if err != nil {
			log.Println(err.Error())
		}
		instanceList = append(instanceList, instance)
	}

	c.JSON(http.StatusOK, struct {
		code int
		msg  string
		data interface{}
	}{10000, "success", instanceList})
}
