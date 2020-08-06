package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/config"
	"github.com/ndphu/skypebot-go/skype/chat"
	"github.com/ndphu/skypebot-go/skype/conversation"
	"github.com/ndphu/skypebot-go/skype/model"
	"github.com/ndphu/skypebot-go/skype/poll"
	"log"
	"strconv"
)

type TokenRequest struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	startServer()
}

func startServer() {
	r := gin.Default()
	manage := r.Group("/api/skype/manage")
	{
		manage.POST("/token", func(c *gin.Context) {
			var token TokenRequest
			if err := c.ShouldBindJSON(&token); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"message": "invalid token request"})
				return
			}
			if err := config.Get().ReloadWithSkypeToken(token.Token); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})
		manage.POST("/login", func(c *gin.Context) {
			var loginRequest LoginRequest
			if err := c.ShouldBindJSON(&loginRequest); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"message": "invalid login request"})
				return
			}
			if token, err := config.Login(loginRequest.Username, loginRequest.Password); err != nil {
				c.AbortWithStatusJSON(500, gin.H{
					"message": "Fail to login. Please check again.",
					"error":   err})
				return
			} else {
				config.Get().ReloadWithSkypeToken(token)
				c.JSON(200, gin.H{"token": token})
			}
		})
	}
	conversations := r.Group("/api/skype/conversations")
	{
		conversations.GET("/", func(c *gin.Context) {
			limit := c.Query("limit")
			if limit == "" {
				limit = "25"
			}
			log.Println("Get conversation list with limit", limit)
			l, err := strconv.Atoi(limit)
			if err != nil {
				l = 25
			}
			res, err := conversation.GetConversations(l)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, res)
			}
		})
	}

	messages := r.Group("/api/skype/messages")
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

	polling := r.Group("/api/skype/polling")
	{
		polling.GET("/endpoints", func(c *gin.Context) {
			if endpoints, err := poll.GetEndpoints(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, endpoints)
			}
		})
	}

	//if config.Get().MessageBaseUrl() != "" {
	//	poll.StartPolling()
	//}

	r.Run()
}
