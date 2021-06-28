package option

import (
	"time"
)

func Conv2Options(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

type Option func(opts *Options)

func WithRequestId(requestId string) Option {
	return func(opts *Options) {
		opts.RequestId = requestId
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opts *Options) {
		opts.Timeout = timeout
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(opts *Options) {
		opts.Headers = headers
	}
}
