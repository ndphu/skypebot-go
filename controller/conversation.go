package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/worker"
	"log"
	"strconv"
)

func Conversations(r *gin.RouterGroup)  {
	conversations := r.Group("/conversations")
	{
		conversations.GET("", func(c *gin.Context) {
			w := worker.FindWorker(c.GetHeader("workerId"))
			if w == nil {
				c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
				return
			}
			limit := c.Query("limit")
			if limit == "" {
				limit = "25"
			}
			log.Println("Get conversation list with limit", limit)
			l, err := strconv.Atoi(limit)
			if err != nil {
				l = 25
			}
			res, err := w.GetConversations(l)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, res)
			}
		})
	}
}
