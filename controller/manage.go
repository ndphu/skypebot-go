package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/manager"
	"github.com/ndphu/skypebot-go/media"
	"github.com/ndphu/skypebot-go/worker"
)

type TokenRequest struct {
	Token string `json:"token"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Manage(r *gin.RouterGroup) {
	manage := r.Group("/manage")
	{
		//manage.POST("/token", func(c *gin.Context) {
		//	w := worker.FindWorker(c.GetHeader("workerId"))
		//	if w == nil {
		//		c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
		//		return
		//	}
		//	var token TokenRequest
		//	if err := c.ShouldBindJSON(&token); err != nil {
		//		c.AbortWithStatusJSON(400, gin.H{"message": "invalid token request"})
		//		return
		//	}
		//	if err := w.Reload(token.Token); err != nil {
		//		c.AbortWithStatusJSON(500, gin.H{"error": err})
		//		return
		//	}
		//	c.JSON(200, gin.H{"success": true})
		//})
		manage.POST("/login", func(c *gin.Context) {
			var loginRequest LoginRequest
			if err := c.ShouldBindJSON(&loginRequest); err != nil {
				c.AbortWithStatusJSON(400, gin.H{"message": "invalid login request"})
				return
			}
			if token, err := worker.Login(loginRequest.Username, loginRequest.Password); err != nil {
				c.AbortWithStatusJSON(500, gin.H{
					"message": "Fail to login. Please check again.",
					"error":   err})
			} else {
				if newWorker, err := worker.NewWorker(token, nil); err != nil {
					c.AbortWithStatusJSON(500, gin.H{
						"message": "Fail to create worker.",
						"error":   err})
				} else {
					manager.AddWorker(newWorker)
					c.JSON(200, gin.H{"success": true, "worker": newWorker.Data()})
				}
			}
		})
		manage.POST("/saveWorkers", func(c *gin.Context) {
			if err := manager.SaveWorkers(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, gin.H{"success": true})
			}
		})

		manage.POST("/reloadMedia", func(c *gin.Context) {
			if err := media.ReloadMedias(); err != nil {
				c.JSON(500, gin.H{"error": err})
			} else {
				c.JSON(200, gin.H{"success": true, "keywords": media.GetKeywords()})
			}
		})
	}
}
