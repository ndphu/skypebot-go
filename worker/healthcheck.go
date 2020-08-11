package worker

import (
	"fmt"
	"github.com/ndphu/skypebot-go/utils"
	"log"
	"time"
)

func (w *Worker) startHealthCheck() {
	if w.healthCheckTicker != nil {
		w.healthCheckTicker.Stop()
	}
	w.healthCheckTicker = time.NewTicker(5 * time.Minute)
	w.sendHealthCheckMessage(fmt.Sprintf("Worker started successfully\n  - ID: %s\n  - User: %s", w.id, w.username))
	go w.healthCheckLoop()
}

func (w *Worker) checkSkypeTokenExp() {
	if w.ShouldRelogin() {
		if w.username != "" && w.password != "" {
			if token, err := Login(w.username, w.password); err == nil {
				w.skypeToken = token
				w.Restart()
			}
		}
	}
}

func (w *Worker) stopHealthCheck() error {
	log.Println("Sending health check stop request...")
	w.stopHealthCheckRequest <- true
	log.Println("Waiting for health check to stop...")
	<-w.stopHealthCheckRequest
	log.Println("Health check stopped successfully")
	return nil
}

func (w *Worker) healthCheckLoop() {
	for {
		select {
		case <-w.healthCheckTicker.C:
			w.checkSkypeTokenExp()
			w.sendHealthCheckMessage("")
			break
		case <-w.stopHealthCheckRequest:
			log.Println("Health check stop request received. Stopping health check timer")
			w.healthCheckTicker.Stop()
			log.Println("Health check timer stopped successfully")
			w.stopHealthCheckRequest <- true
			break
		}
	}
}

func (w *Worker) sendHealthCheckMessage(message string) {
	if w.healthCheckThread != "" {
		if message == "" {
			message = time.Now().Format(time.RFC3339)
		}
		if err := utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(w.healthCheckThread, message)
		}); err != nil {
			log.Println("Fail to send health check message", err.Error())
		}
	}
}
