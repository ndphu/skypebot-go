package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/worker"
	"log"
)

type PollingRequest struct {
	Token string `json:"token"`
}

func Polling(r *gin.RouterGroup) {
	pollingGroup := r.Group("/polling")
	{

		pollingGroup.POST("/startPolling", func(c *gin.Context) {
			request := PollingRequest{}
			if err := c.ShouldBindJSON(&request); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": "invalid polling request"})
				return
			}
			pollingWorker, err := worker.NewWorker(request.Token, func(event *worker.EventMessage) {
				log.Println(event.Resource.Content)
			})
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "Fail to create polling worker", "error": err})
				return
			}
			if err := pollingWorker.Start(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"message": "Fail to start polling worker", "error": err})
				return
			}
			c.JSON(200, gin.H{"status": "success"})
		})
	}
}
