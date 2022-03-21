package router

import (
	"fmt"
	"log"
	"net/http"
	"raptor/jobcenter"
	"raptor/proto"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Route(router *gin.Engine) {
	router.POST("/register", registerJob)
	router.GET("/jobData", getJobData)

	scheduling(router.Group("/scheduler"))
}

func scheduling(router *gin.RouterGroup) {
	jobcenter.New()
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})
	router.GET("/timing", timingJob)
	//更新从节点任务状态
	router.GET("/jobstate", updateJobState)
}

func registerJob(c *gin.Context) {
	jobRegister := jobcenter.New()

	var body proto.Config
	c.ShouldBind(&body)
	if err := jobRegister.Register(body); err != nil {
		log.Println(err.Error())
		c.String(http.StatusOK, "register failed")
		return
	}

	c.JSON(http.StatusOK, "register success")
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
	fmt.Println(jobInstanceID, state)
}
