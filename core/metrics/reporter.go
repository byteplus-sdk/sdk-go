package metrics

import (
	"time"
)

type Reporter struct {
	param   *ReporterBuilder
	manager *Manager
}

type ReporterBuilder struct {
	enableMetrics bool
	prefix        string // prefix of all metrics name
	baseTags      map[string]string
}

func NewReporterBuilder() *ReporterBuilder {
	return &ReporterBuilder{}
}

func (b *ReporterBuilder) EnableMetrics(enableMetrics bool) *ReporterBuilder {
	b.enableMetrics = enableMetrics
	return b
}

func (b *ReporterBuilder) BaseTags(baseTags map[string]string) *ReporterBuilder {
	if len(baseTags) != 0 {
		b.baseTags = baseTags
	}
	return b
}

func (b *ReporterBuilder) Prefix(prefix string) *ReporterBuilder {
	if prefix != "" {
		b.prefix = prefix
	}
	return b
}

func (b *ReporterBuilder) Build() *Reporter {
	if b.baseTags == nil {
		b.baseTags = make(map[string]string)
	}
	b.baseTags["host"] = getLocalHost()
	if b.prefix == "" {
		b.prefix = defaultMetricsPrefix
	}
	return &Reporter{
		param:   b,
		manager: GetManager(b.prefix),
	}
}

func (r *Reporter) stop() {
	r.manager.Stop()
}

/**
 * Counter介绍：https://site.bytedance.net/docs/2080/2717/36906/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Counter("request.count", 1, "method:user")
 */

func (r *Reporter) Counter(key string, value float64, tagKvs ...string) {
	if !r.param.enableMetrics {
		return
	}
	r.manager.emitCounter(key, appendTags(r.param.baseTags, tagKvs), value)
}

/**
 * Timer介绍：https://site.bytedance.net/docs/2080/2717/36907/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Timer("request.cost", 100, "method:user")
 */

func (r *Reporter) Timer(key string, value float64, tagKvs ...string) {
	if !r.param.enableMetrics {
		return
	}
	r.manager.emitTimer(key, appendTags(r.param.baseTags, tagKvs), value)
}

/**
 * Latency介绍：Latency基于timer封装，非Metrics的标准类型。timer介绍：https://site.bytedance.net/docs/2080/2717/36907/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Latency("request.cost", startTime, "method:user")
 */

func (r *Reporter) Latency(key string, begin time.Time, tagKvs ...string) {
	if !r.param.enableMetrics {
		return
	}
	r.manager.emitTimer(key, appendTags(r.param.baseTags, tagKvs), float64(time.Now().Sub(begin).Milliseconds()))
}

/**
 * Store介绍：https://site.bytedance.net/docs/2080/2717/36905/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Store("goroutine.count", 400, "ip:127.0.0.1")
 */

func (r *Reporter) Store(key string, value float64, tagKvs ...string) {
	if !r.param.enableMetrics {
		return
	}
	r.manager.emitStore(key, appendTags(r.param.baseTags, tagKvs), value)
}
