package utils

import "time"

type RetryParams struct {
	Retry         int           `json:"retry"`
	SleepInterval time.Duration `json:"sleepInterval"`
}

func ExecuteWithRetryTimes(function func() error, params RetryParams) error {
	if params.SleepInterval == 0 {
		params.SleepInterval = 2 * time.Second
	}
	try := 0
	var postError error
	for ; try < params.Retry || params.Retry == 0; {
		if postError = function(); postError != nil {
			time.Sleep(params.SleepInterval)
		} else {
			return nil
		}
		try ++
	}
	return postError
}

func ExecuteWithRetry(function func() error) error {
	return ExecuteWithRetryTimes(function, RetryParams{
		Retry:         5,
		SleepInterval: 2 * time.Second,
	})
}
