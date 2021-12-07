package metrics

import "time"

const (
	defaultMetricsDomain = "bot.snssdk.com"
	defaultMetricsPrefix = "byteplus.rec.sdk"

	counterUrlFormat = "http://%s/api/counter"
	otherUrlFormat   = "http://%s/api/put"

	//max number of metrics to be processed for each flush
	maxFlashSize = 65536 * 2
	//default tidy interval
	defaultTidyInterval = 100 * time.Second
	//interval of flushing all cache metrics
	defaultFlushInterval = 15 * time.Second
	//default expire interval of each counter/timer/store, expired metrics will be cleaned
    defaultMetricsExpireTime = 12 * time.Hour
)

type metricsType int

const (
	metricsTypeCounter metricsType = iota
	metricsTypeTimer
	metricsTypeStore
)
