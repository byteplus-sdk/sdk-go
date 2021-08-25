package option

import "time"

type Options struct {
	Timeout       time.Duration
	RequestId     string
	Headers       map[string]string
	DataDate      time.Time
	DataIsEnd     bool
	ServerTimeout time.Duration
	Queries       map[string]string
	Stage         string
}
