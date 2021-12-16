package option

import "time"

type Options struct {
	Timeout       time.Duration
	RequestId     string
	Headers       map[string]string
	DataDate      time.Time
	DataIsEnd     bool
	ServerTimeout time.Duration
	Stage         string
	Queries       map[string]string
	Scene         string
}
