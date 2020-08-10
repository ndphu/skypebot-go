package worker

import (
	"github.com/ndphu/skypebot-go/utils"
	"time"
)

func (w *Worker) startHealthCheck() {
	go w.checkSkypeTokenExp()

	if w.healthCheckThread != "" {
		go w.doHealthCheck()
	}
}

func (w*Worker) checkSkypeTokenExp()  {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <- ticker.C:
			if w.ShouldRelogin() {
				if w.username != "" && w.password != "" {
					if token, err := Login(w.username, w.password); err == nil {
						w.skypeToken = token
						w.Restart()
					}

				}
			}
		}
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
