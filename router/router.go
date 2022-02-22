package router

import (
	"log"
	"net/http"
	"raptor/jobcenter"
	"raptor/proto"

	"github.com/gin-gonic/gin"
)

func Route(router *gin.Engine) {
	router.POST("/register", registerJob)
	router.GET("/jobData", getJobData)

	scheduling(router.Group("/scheduler"))
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

func scheduling(router *gin.RouterGroup) {
	router.POST("/timing", timingJob)
}

func timingJob(c *gin.Context) {
	scheduler := jobcenter.New()

	isMaster := c.Query("isMaster")
	jobID := c.Query("id")

	scheduler.TimingJob(isMaster, jobID)

}
