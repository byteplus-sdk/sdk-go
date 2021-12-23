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
	return func(options *Options) {
		options.RequestId = requestId
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		options.Timeout = timeout
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(options *Options) {
		options.Headers = headers
	}
}

func WithDataDate(date time.Time) Option {
	return func(options *Options) {
		options.DataDate = date
	}
}

func WithDateEnd(isEnd bool) Option {
	return func(options *Options) {
		options.DataIsEnd = isEnd
	}
}

func WithServerTimeout(timeout time.Duration) Option {
	return func(options *Options) {
		options.ServerTimeout = timeout
	}
}

func WithStage(stage string) Option {
	return func(options *Options) {
		options.Stage = stage
	}
}

func WithQueries(queries map[string]string) Option {
	return func(options *Options) {
		options.Queries = queries
	}
}

func WithScene(scene string) Option {
	return func(options *Options) {
		options.Scene = scene
	}
}
