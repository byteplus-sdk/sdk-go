package metrics

import "time"

const (
	defaultMetricsDomain = "bot.snssdk.com"
	defaultMetricsPrefix = "byteplus.rec.sdk"

	counterUrlFormat = "https://%s/api/counter"
	otherUrlFormat   = "https://%s/api/put"

	defaultFlushInterval = 10 * time.Second
	reservoirSize        = 65536
	decayAlpha           = 0.02
	maxTryTimes          = 2
	defaultHttpTimeout   = 800 * time.Millisecond

	delimiter = "+"
)

type metricsType int

const (
	metricsTypeCounter metricsType = iota
	metricsTypeTimer
	metricsTypeStore
)
