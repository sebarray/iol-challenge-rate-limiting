package entity

import "time"

type Bucket struct {
	Tokens     float64
	LastRefill time.Time
	LastSeen   time.Time
}
