package worker

import (
	"github.com/ndphu/skypebot-go/utils"
	"time"
)

func (w *Worker) startHealthCheck() {
	if w.healthCheckThread != "" {
		go w.doHealthCheck()
	}
}

func (w *Worker) doHealthCheck() {
	for ; ; {
		if err := utils.ExecuteWithRetry(func() error {
			return w.SendTextMessage(w.healthCheckThread, time.Now().Format(time.RFC3339))
		}); err != nil {
			break
		}
		time.Sleep(5 *  time.Minute)
	}
}
