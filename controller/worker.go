package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/manager"
	"github.com/ndphu/skypebot-go/model"
	"github.com/ndphu/skypebot-go/worker"
)

func WorkerController(r *gin.RouterGroup)  {

	r.GET("/workers", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"success": true,
			"workers": manager.GetWorkers(),
		})
	})

	r.POST("/worker/:workerId/start", func(c *gin.Context) {
		w := manager.FindWorker(c.Param("workerId"))
		if w == nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
			return
		}
		if w.Data().Status != worker.StatusRunning &&
			w.Data().Status != worker.StatusStarting {
			if err := w.Start(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "fail to start worker", "worker": w.Data()})
				return
			}
		}
		c.JSON(200, gin.H{"success": true, "worker": w.Data()})
	})
	r.POST("/worker/:workerId/stop", func(c *gin.Context) {
		w := manager.FindWorker(c.Param("workerId"))
		if w == nil {
			c.AbortWithStatusJSON(404, gin.H{"error": "worker not found"})
			return
		}
		if w.Data().Status != worker.StatusStopping&&
			w.Data().Status != worker.StatusStopped {
			if err := w.Stop(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "fail to stop worker", "worker": w.Data()})
				return
			}
		}
		c.JSON(200, gin.H{"success": true, "worker": w.Data()})
	})


	r.POST("/worker/:workerId/message/text", func(c *gin.Context) {
		w := manager.FindWorker(c.Param("workerId"))
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


}
