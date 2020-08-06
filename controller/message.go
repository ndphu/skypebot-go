package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/skype/chat"
	"github.com/ndphu/skypebot-go/skype/model"
	"log"
)

func Messages(r * gin.RouterGroup)  {
	messages := r.Group("/messages")
	{

		messages.POST("/textMessage", func(c *gin.Context) {
			req := model.PostTextMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := chat.PostTextMessage(req.Target, req.Text); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.POST("/reactMessage", func(c *gin.Context) {
			req := model.ReactMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := chat.ReactMessage(req.Target, req.MessageId, req.Emotion); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.POST("/reactThread", func(c *gin.Context) {
			req := model.ReactMessageRequest{}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
			if err := chat.ReactThread(req.Target, req.Emotion); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})

		messages.GET("/thread/:threadId", func(c *gin.Context) {
			target := c.Param("threadId")
			log.Println("Load message from thread:", target)
			if existingMessages, err := chat.GetMessages(target); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			} else {
				c.JSON(200, existingMessages)
			}
		})
	}
}