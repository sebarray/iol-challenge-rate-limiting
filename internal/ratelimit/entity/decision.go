package entity

import "time"

type Decision struct {
	Allowed    bool
	Limit      int
	Remaining  int
	RetryAfter time.Duration
	ResetAfter time.Duration
}
