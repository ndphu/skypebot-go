package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/worker"
	"log"
)

func Messages(r * gin.RouterGroup)  {
	messages := r.Group("/messages")
	{
		messages.POST("/textMessage", func(c *gin.Context) {
			w := worker.FindWorker(c.GetHeader("workerId"))
			if w == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
				return
			}
			req := model.PostTextMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := w.SendTextMessage(req.Target, req.Text); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.POST("/reactMessage", func(c *gin.Context) {
			w := worker.FindWorker(c.GetHeader("workerId"))
			if w == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
				return
			}
			req := model.ReactMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := w.ReactMessage(req.Target, req.MessageId, req.Emotion); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.POST("/reactThread", func(c *gin.Context) {
			w := worker.FindWorker(c.GetHeader("workerId"))
			if w == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
				return
			}
			req := model.ReactMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := w.ReactThread(req.Target, req.Emotion); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.GET("/thread/:threadId", func(c *gin.Context) {
			w := worker.FindWorker(c.GetHeader("workerId"))
			if w == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
				return
			}
			target := c.Param("threadId")
			log.Println("Load message from thread:", target)
			if existingMessages, err := w.GetMessages(target); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			} else {
				c.JSON(200, existingMessages)
			}
		})
	}
}