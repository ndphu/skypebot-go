package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/config"
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
}