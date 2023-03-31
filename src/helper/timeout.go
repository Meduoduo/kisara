package helper

import (
	"time"
)

// TimeoutWrapper is a wrapper for timeout, which will return true if timeout, if not, return false
// success_chan is required to be passed to infer that the task is done
func TimeoutWrapper(timeout time.Duration, success_chan chan bool) bool {
	timer := time.NewTimer(timeout)
	select {
	case <-timer.C:
		return false
	case <-success_chan:
		return true
	}
}
