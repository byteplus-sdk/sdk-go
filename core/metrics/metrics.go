package metrics

import "time"

type Metrics interface {
	getName() string
	isExpired() bool
	updateExpireTime(ttl time.Duration)
	flush()
	emit(float64, map[string]string)
}
