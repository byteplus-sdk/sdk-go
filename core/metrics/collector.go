package metrics

import (
	"errors"
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/valyala/fasthttp"
	"go.uber.org/atomic"
	"google.golang.org/protobuf/proto"
	"math"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

var (
	metricsCfg       *config
	metricsCollector *collector
	onceLock         = &sync.Once{}
)

type collector struct {
	collectors map[metricsType]map[string]*metricValue
	locks      map[metricsType]*sync.RWMutex
	httpCli    *fasthttp.Client
}

type config struct {
	enableMetrics bool
	domain        string
	prefix        string
	printLog      bool // whether print logs during collecting metrics
	flushInterval time.Duration
}

type metricValue struct {
	value        interface{}
	flushedValue interface{}
	updated      bool // When there is a new report, updated is true, otherwise it is false
}

type timerValue struct {
	sample Sample
	count  *atomic.Int64
}

// Init As long as the Init function is called, the metrics are enabled
func Init(options ...Option) {
	// if no options, set to default config
	metricsCfg = &config{
		domain:        defaultMetricsDomain,
		flushInterval: defaultFlushInterval,
		prefix:        defaultMetricsPrefix,
		enableMetrics: true,
	}
	for _, option := range options {
		option(metricsCfg)
	}
	metricsCollector = &collector{
		httpCli: &fasthttp.Client{},
		collectors: map[metricsType]map[string]*metricValue{
			metricsTypeCounter: make(map[string]*metricValue),
			metricsTypeTimer:   make(map[string]*metricValue),
			metricsTypeStore:   make(map[string]*metricValue),
		},
		locks: map[metricsType]*sync.RWMutex{
			metricsTypeCounter: {},
			metricsTypeStore:   {},
			metricsTypeTimer:   {},
		},
	}
	onceLock.Do(func() {
		startReport()
	})
}

type Option func(*config)

func WithMetricsDomain(domain string) Option {
	return func(config *config) {
		if domain != "" {
			config.domain = domain
		}
	}
}

func WithMetricsPrefix(prefix string) Option {
	return func(config *config) {
		if prefix != "" {
			config.prefix = prefix
		}
	}
}

//WithMetricsLog if not set, will not print metrics log
func WithMetricsLog() Option {
	return func(config *config) {
		config.printLog = true
	}
}

// WithFlushInterval set the interval of reporting metrics
func WithFlushInterval(flushInterval time.Duration) Option {
	return func(config *config) {
		if flushInterval > 500*time.Millisecond { // flushInterval should not be too small
			config.flushInterval = flushInterval
		}
	}
}

func isEnableMetrics() bool {
	if metricsCfg == nil {
		return false
	}
	return metricsCfg.enableMetrics
}

// isEnablePrintLog enable print log during reporting metrics
func isEnablePrintLog() bool {
	if metricsCfg == nil {
		return false
	}
	return metricsCfg.printLog
}

// 更新(name,tags)对应的value为最新值，每次上报完无需清空
func emitStore(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	getOrStoreDefaultMetric(metricsTypeStore, collectKey)
	updateMetric(metricsTypeStore, collectKey, value)
}

// 统计本次上报期间（flushInterval）内(name,tags)对应value的累加值，每次上报完需清空
func emitCounter(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	getOrStoreDefaultMetric(metricsTypeCounter, collectKey)
	updateMetric(metricsTypeCounter, collectKey, value)
}

func emitTimer(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	getOrStoreDefaultMetric(metricsTypeTimer, collectKey)
	updateMetric(metricsTypeTimer, collectKey, value)
}

func getOrStoreDefaultMetric(metricType metricsType, collectKey string) {
	if isMetricExist(metricType, collectKey) {
		return
	}
	metricsCollector.locks[metricType].Lock()
	defer metricsCollector.locks[metricType].Unlock()
	if metricsCollector.collectors[metricType][collectKey] == nil {
		metricsCollector.collectors[metricType][collectKey] = buildDefaultMetric(metricType)
	}
}

func isMetricExist(metricType metricsType, collectKey string) bool {
	metricsCollector.locks[metricType].RLock()
	defer metricsCollector.locks[metricType].RUnlock()
	if metricsCollector.collectors[metricType][collectKey] != nil {
		return true
	}
	return false
}

func buildDefaultMetric(metricType metricsType) *metricValue {
	if metricType == metricsTypeTimer {
		return &metricValue{
			value: &timerValue{
				sample: NewExpDecaySample(reservoirSize, decayAlpha),
				count:  atomic.NewInt64(0),
			},
			updated: false,
		}
	}
	return &metricValue{
		value:        atomic.NewFloat64(0),
		flushedValue: atomic.NewFloat64(0),
		updated:      false,
	}
}

func updateMetric(metricType metricsType, collectKey string, value float64) {
	metricsCollector.locks[metricType].RLock()
	defer metricsCollector.locks[metricType].RUnlock()
	oldValue := metricsCollector.collectors[metricType][collectKey]
	switch metricType {
	case metricsTypeStore:
		oldValue.value.(*atomic.Float64).Store(value)
	case metricsTypeCounter:
		oldValue.value.(*atomic.Float64).Add(value)
	case metricsTypeTimer:
		oldValue.value.(*timerValue).sample.Update(int64(value))
		oldValue.value.(*timerValue).count.Inc()
	}
	oldValue.updated = true
}

func startReport() {
	if !isEnableMetrics() {
		return
	}
	ticker := time.NewTicker(metricsCfg.flushInterval)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if isEnablePrintLog() {
					logs.Error("metrics report encounter panic:%+v, stack:%s", err, string(debug.Stack()))
				}
			}
		}()
		for range ticker.C {
			flushTimer()
			flushStore()
			flushCounter()
		}
	}()
}

func flushStore() {
	metricsRequests := make([]*Metric, 0)
	metricsCollector.locks[metricsTypeStore].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeStore] {
		name, tagKvs, ok := parseNameAndTags(key)
		if !ok {
			continue
		}
		if metric.updated { // if updated is false, means no metric emit
			metricsRequest := &Metric{
				Metric:    metricsCfg.prefix + "." + name,
				Tags:      tagKvs,
				Value:     metric.value.(*atomic.Float64).Load(),
				Timestamp: uint64(time.Now().Unix()),
			}
			metricsRequests = append(metricsRequests, metricsRequest)
			// reset updated tag after report
			metric.updated = false
			metric.value.(*atomic.Float64).Store(0)
		}
	}
	metricsCollector.locks[metricsTypeStore].RUnlock()
	if len(metricsRequests) > 0 {
		url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
		if err := send(&MetricMessage{Metrics: metricsRequests}, url); err != nil {
			if isEnablePrintLog() {
				logs.Error("[Metrics] exec store err:%+v, url:%s, metricsRequests:%+v", err, url, metricsRequests)
			}
			return
		}
		if isEnablePrintLog() {
			logs.Debug("[Metrics] exec store success, url:%s, metricsRequests:%+v", url, metricsRequests)
		}
	}
}

func flushCounter() {
	metricsRequests := make([]*Metric, 0)
	metricsCollector.locks[metricsTypeCounter].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeCounter] {
		name, tagKvs, ok := parseNameAndTags(key)
		if !ok {
			continue
		}
		if metric.updated {
			valueCopy := metric.value.(*atomic.Float64).Load()
			metricsRequest := &Metric{
				Metric:    metricsCfg.prefix + "." + name,
				Tags:      tagKvs,
				Value:     valueCopy - metric.flushedValue.(*atomic.Float64).Load(),
				Timestamp: uint64(time.Now().Unix()),
			}
			metricsRequests = append(metricsRequests, metricsRequest)
			// reset updated tag after report
			metric.updated = false
			// after each flushInterval of the counter is reported, the accumulated metric needs to be cleared
			metric.flushedValue.(*atomic.Float64).Store(valueCopy)
			// if the value is too large, reset it
			if valueCopy >= math.MaxFloat64/2 {
				metric.value.(*atomic.Float64).Store(0)
				metric.flushedValue.(*atomic.Float64).Store(0)
			}
		}
	}
	metricsCollector.locks[metricsTypeCounter].RUnlock()

	if len(metricsRequests) > 0 {
		url := fmt.Sprintf(counterUrlFormat, metricsCfg.domain)
		if err := send(&MetricMessage{Metrics: metricsRequests}, url); err != nil {
			if isEnablePrintLog() {
				logs.Error("[Metrics] exec counter err:%+v, url:%s, metricsRequests:%+v", err, url, metricsRequests)
			}
			return
		}
		if isEnablePrintLog() {
			logs.Debug("[Metrics] exec counter success, url:%s, metricsRequests:%+v", url, metricsRequests)
		}
	}
}

func flushTimer() {
	metricsRequests := make([]*Metric, 0)
	metricsCollector.locks[metricsTypeTimer].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeTimer] {
		name, tagKvs, ok := parseNameAndTags(key)
		if !ok {
			return
		}
		if metric.updated {
			timestamp := time.Now().Unix()
			snapshot := metric.value.(*timerValue).sample.Snapshot()
			count := metric.value.(*timerValue).count.Load()

			//qps
			//The timerValue data will be transmitted to metrics_proxy and reported as store metric,
			//so the qps here must also be store metric, which need to be divided by flushInterval,
			//and the qps metric should be showed as store in grafana.
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "qps",
				Tags:      tagKvs,
				Value:     float64(count) / metricsCfg.flushInterval.Seconds(),
				Timestamp: uint64(timestamp),
			})
			//max
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "max",
				Tags:      tagKvs,
				Value:     float64(snapshot.Max()),
				Timestamp: uint64(timestamp),
			})
			//min
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "min",
				Tags:      tagKvs,
				Value:     float64(snapshot.Min()),
				Timestamp: uint64(timestamp),
			})
			//avg
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "avg",
				Tags:      tagKvs,
				Value:     snapshot.Mean(),
				Timestamp: uint64(timestamp),
			})
			//pc75
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "pct75",
				Tags:      tagKvs,
				Value:     snapshot.Percentile(0.75),
				Timestamp: uint64(timestamp),
			})
			//pc90
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "pct90",
				Tags:      tagKvs,
				Value:     snapshot.Percentile(0.90),
				Timestamp: uint64(timestamp),
			})
			//pc95
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "pct95",
				Tags:      tagKvs,
				Value:     snapshot.Percentile(0.95),
				Timestamp: uint64(timestamp),
			})
			//pc99
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "pct99",
				Tags:      tagKvs,
				Value:     snapshot.Percentile(0.99),
				Timestamp: uint64(timestamp),
			})
			//pc999
			metricsRequests = append(metricsRequests, &Metric{
				Metric:    metricsCfg.prefix + "." + name + "." + "pct999",
				Tags:      tagKvs,
				Value:     snapshot.Percentile(0.999),
				Timestamp: uint64(timestamp),
			})

			// reset updated tag after report
			metric.updated = false
			//count is the counter metric and needs to be cleared after each report
			metric.value.(*timerValue).count.Store(0)
		}
	}
	metricsCollector.locks[metricsTypeTimer].RUnlock()
	if len(metricsRequests) > 0 {
		url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
		if err := send(&MetricMessage{Metrics: metricsRequests}, url); err != nil {
			if isEnablePrintLog() {
				logs.Error("[Metrics] exec timer err:%+v, url:%s, metricsRequests:%+v", err, url, metricsRequests)
			}
			return
		}
		if isEnablePrintLog() {
			logs.Debug("[Metrics] exec timer success, url:%s, metricsRequests:%+v", url, metricsRequests)
		}
	}
}

// send httpRequest to metrics server
func send(metricRequests *MetricMessage, url string) error {
	var err error
	var request *fasthttp.Request
	for i := 0; i < maxTryTimes; i++ {
		request, err = buildMetricsRequest(metricRequests, url)
		if err != nil {
			fasthttp.ReleaseRequest(request)
			continue
		}
		err = doSend(request)
		if err == nil {
			return nil
		}
		// retry when http timeout
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			err = errors.New("request timeout, msg:" + err.Error())
			continue
		}
		// when occur other err, return directly
		return err
	}
	return err
}

func doSend(request *fasthttp.Request) error {
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	err := metricsCollector.httpCli.DoTimeout(request, response, defaultHttpTimeout)
	if err == nil && response.StatusCode() == fasthttp.StatusOK {
		return nil
	}
	return err
}

func buildMetricsRequest(metricRequests *MetricMessage, url string) (*fasthttp.Request, error) {
	request := fasthttp.AcquireRequest()
	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(url)
	//request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Type", "application/protobuf")
	request.Header.Set("Accept", "application/json")
	//body, err := json.Marshal(metricRequests)
	body, err := proto.Marshal(metricRequests)

	if err != nil {
		return nil, err
	}
	request.SetBodyRaw(body)
	return request, nil
}
