package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/controller"
	"github.com/ndphu/skypebot-go/manager"
	"github.com/ndphu/skypebot-go/media"
)


func main() {
	startServer()
}

func startServer() {
	media.ReloadMedias()
	manager.Start()

	r := gin.Default()

	skypeEndpoint := r.Group("/api/skype")

	controller.Manage(skypeEndpoint)
	controller.Conversations(skypeEndpoint)
	controller.Messages(skypeEndpoint)
	controller.WorkerController(skypeEndpoint)

	r.Run()
}
