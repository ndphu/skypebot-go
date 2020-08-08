package worker

import "time"

func (w *Worker) startHealthCheck() {
	if w.healthCheckThread != "" {
		go w.doHealthCheck()
	}
}

func (w *Worker) doHealthCheck() {
	for ; ; {
		if err := w.postWithRetry(func() error {
			return w.PostTextMessage(w.healthCheckThread, time.Now().Format(time.RFC3339))
		}, 3, 2*time.Second); err != nil {
			break
		}
		time.Sleep(5 *  time.Minute)
	}
}
