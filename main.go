package main

import (
	"github.com/gin-gonic/gin"
	"github.com/ndphu/skypebot-go/controller"
)


func main() {
	startServer()
}

func startServer() {
	r := gin.Default()

	skypeEndpoint := r.Group("/api/skype")

	controller.Manage(skypeEndpoint)
	controller.Conversations(skypeEndpoint)
	controller.Messages(skypeEndpoint)
	controller.WorkerController(skypeEndpoint)

	r.Run()
}
