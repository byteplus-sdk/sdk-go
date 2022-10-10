package metrics

import "time"

const (
	// metrics default domain and prefix
	defaultMetricsDomain     = "rec-api-sg1.recplusapi.com"
	defaultMetricsPrefix     = "byteplus.rec.sdk"
	defaultMetricsHTTPSchema = "https"

	// monitor url format
	metricsURLFormat    = "%s://%s/predict/api/monitor/metrics"
	metricsLogURLFormat = "%s://%s/predict/api/monitor/metrics/log"

	// domain path
	metricsPath    = "/monitor/metrics"
	metricsLogPath = "/monitor/metrics/log"

	// metrics base config
	defaultReportInterval = 15 * time.Second
	defaultHTTPTimeout    = 800 * time.Millisecond
	maxTryTimes           = 3
	maxSpinTimes          = 5
	successHTTPCode       = 200
	maxMetricsSize        = 10000
	maxMetricsLogSize     = 5000

	// metrics log level
	logLevelTrace  = "trace"
	logLevelDebug  = "debug"
	logLevelInfo   = "info"
	logLevelNotice = "notice"
	logLevelWarn   = "warn"
	logLevelError  = "error"
	logLevelFatal  = "fatal"

	// metrics type
	metricsTypeCounter     = "counter"
	metricsTypeStore       = "store"
	metricsTypeTimer       = "timer"
	metricsTypeRateCounter = "rate_counter"
	metricsTypeMeter       = "meter"
)
