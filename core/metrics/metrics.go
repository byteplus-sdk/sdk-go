package metrics

// Store description: Store tagKvs should be formatted as "key:value"
// example: store("goroutine.count", 400, "ip:127.0.0.1")
func Store(key string, value int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeStore, key, value, tagKvs...)
}

// Counter description: Store tagKvs should be formatted as "key:value"
// example: counter("request.count", 1, "method:user", "type:upload")
func Counter(key string, value int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeCounter, key, value, tagKvs...)
}

// Timer The unit of `value` is milliseconds
// example: timer("request.cost", 100, "method:user", "type:upload")
// description: Store tagKvs should be formatted as "key:value"
func Timer(key string, value int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeTimer, key, value, tagKvs...)
}

// Latency The unit of `begin` is milliseconds
// example: latency("request.latency", startTime, "method:user", "type:upload")
// description: Store tagKvs should be formatted as "key:value"
func Latency(key string, begin int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeTimer, key, currentTimeMillis()-begin, tagKvs...)
}

// RateCounter description: Store tagKvs should be formatted as "key:value"
// example: rateCounter("request.count", 1, "method:user", "type:upload")
func RateCounter(key string, value int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeRateCounter, key, value, tagKvs...)
}

// Meter description:
//  - meter(xx) = counter(xx) + rateCounter(xx.rate)
//  - Store tagKvs should be formatted as "key:value"
// example: rateCounter("request.count", 1, "method:user", "type:upload")
func Meter(key string, value int64, tagKvs ...string) {
	Collector.EmitMetric(metricsTypeMeter, key, value, tagKvs...)
}
