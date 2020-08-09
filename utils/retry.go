package utils

import "time"

type RetryParams struct {
	Retry         int           `json:"retry"`
	SleepInterval time.Duration `json:"sleepInterval"`
}

func ExcuteWithRetryTimes(function func()error, params RetryParams) error {
	if params.Retry == 0 {
		params.Retry = 5
	}
	if params.SleepInterval == 0 {
		params.SleepInterval = 2 * time.Second
	}
	try := 0
	var postError error
	for ; try < params.Retry; {
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
		return ExcuteWithRetryTimes(function, RetryParams{
			Retry: 5,
			SleepInterval: 2 * time.Second,
		})
}
