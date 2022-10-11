package metrics

import (
	"time"
)

type Option func(config *Config)

func WithMetricsDomain(domain string) Option {
	return func(config *Config) {
		if domain != "" {
			config.Domain = domain
		}
	}
}

func WithMetricsPrefix(prefix string) Option {
	return func(config *Config) {
		if prefix != "" {
			config.Prefix = prefix
		}
	}
}

func WithMetricsHTTPSchema(schema string) Option {
	return func(config *Config) {
		if schema == "http" || schema == "https" {
			config.HTTPSchema = schema
		}
	}
}

// EnableMetrics if not set, will not report metrics.
func EnableMetrics() Option {
	return func(config *Config) {
		config.EnableMetrics = true
	}
}

// EnableMetricsLog if not set, will not report metrics logs.
func EnableMetricsLog() Option {
	return func(config *Config) {
		config.EnableMetricsLog = true
	}
}

// WithReportInterval set the interval of reporting metrics
func WithReportInterval(reportInterval time.Duration) Option {
	return func(config *Config) {
		// reportInterval should not be too small
		if reportInterval.Milliseconds() > 1000 {
			config.ReportInterval = reportInterval
		}
	}
}

func WithMetricsTimeout(timeout time.Duration) Option {
	return func(config *Config) {
		config.HTTPTimeout = timeout
	}
}
