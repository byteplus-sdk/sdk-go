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
	// timer stat names to be reported
	timerStatMetrics = []string{"qps", "max", "min", "avg", "pct75", "pct90", "pct95", "pct99", "pct999"}
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
	updateMetric(metricsTypeStore, collectKey, value)
}

// 统计本次上报期间（flushInterval）内(name,tags)对应value的累加值，每次上报完需清空
func emitCounter(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	updateMetric(metricsTypeCounter, collectKey, value)
}

func emitTimer(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	updateMetric(metricsTypeTimer, collectKey, value)
}

func updateMetric(metricType metricsType, collectKey string, value float64) {
	setDefaultMetricIfNotExist(metricType, collectKey)
	metricsCollector.locks[metricType].RLock()
	defer metricsCollector.locks[metricType].RUnlock()
	oldValue := metricsCollector.collectors[metricType][collectKey]
	switch metricType {
	case metricsTypeStore:
		oldValue.value = value
	case metricsTypeCounter:
		oldValue.value.(*atomic.Float64).Add(value)
	case metricsTypeTimer:
		oldValue.value.(Sample).Update(int64(value))
	}
	oldValue.updated = true
}

func setDefaultMetricIfNotExist(metricType metricsType, collectKey string) {
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
	switch metricType {
	//timer
	case metricsTypeTimer:
		return &metricValue{
			value:   NewUniformSample(reservoirSize),
			updated: false,
		}
	//counter
	case metricsTypeCounter:
		return &metricValue{
			value:        atomic.NewFloat64(0),
			flushedValue: atomic.NewFloat64(0),
			updated:      false,
		}
	}
	//store
	return &metricValue{
		value:   float64(0),
		updated: false,
	}
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
	metricsRequests := make([]*Metric, 0, len(metricsCollector.collectors[metricsTypeStore]))
	metricsCollector.locks[metricsTypeStore].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeStore] {
		if metric.updated { // if updated is false, means no metric emit
			name, tagKvs, ok := parseNameAndTags(key)
			if !ok {
				continue
			}
			metricsRequest := &Metric{
				Metric:    metricsCfg.prefix + "." + name,
				Tags:      tagKvs,
				Value:     metric.value.(float64),
				Timestamp: uint64(time.Now().Unix()),
			}
			metricsRequests = append(metricsRequests, metricsRequest)
			// reset updated tag after report
			metric.updated = false
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
	metricsRequests := make([]*Metric, 0, len(metricsCollector.collectors[metricsTypeCounter]))
	metricsCollector.locks[metricsTypeCounter].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeCounter] {
		if metric.updated {
			name, tagKvs, ok := parseNameAndTags(key)
			if !ok {
				continue
			}
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
	metricsRequests := make([]*Metric, 0, len(metricsCollector.collectors[metricsTypeTimer])*len(timerStatMetrics))
	metricsCollector.locks[metricsTypeTimer].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeTimer] {
		if metric.updated {
			name, tagKvs, ok := parseNameAndTags(key)
			if !ok {
				return
			}
			snapshot := metric.value.(Sample).Snapshot()
			metricsRequests = append(metricsRequests, buildStatMetrics(snapshot, name, tagKvs)...)
			// reset updated tag after report
			metric.updated = false
			// clear sample every sample period
			metric.value.(Sample).Clear()
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

func buildStatMetrics(sample Sample, name string, tagKvs map[string]string) []*Metric {
	timestamp := uint64(time.Now().Unix())
	metricsRequests := make([]*Metric, 0, len(timerStatMetrics))
	for _, statName := range timerStatMetrics {
		metricsRequests = append(metricsRequests, &Metric{
			Metric:    metricsCfg.prefix + "." + name + "." + statName,
			Tags:      tagKvs,
			Value:     getStatValue(statName, sample),
			Timestamp: timestamp,
		})
	}
	return metricsRequests
}

func getStatValue(statName string, sample Sample) float64 {
	switch statName {
	//qps
	//The timerValue data will be transmitted to metrics_proxy and reported as store metric,
	//so the qps here must also be store metric, which need to be divided by flushInterval,
	//and the qps metric should be showed as store in grafana.
	case "qps":
		return float64(sample.Count()) / metricsCfg.flushInterval.Seconds()
	case "max":
		return float64(sample.Max())
	case "min":
		return float64(sample.Min())
	case "avg":
		return sample.Mean()
	case "pct75":
		return sample.Percentile(0.75)
	case "pct90":
		return sample.Percentile(0.90)
	case "pct95":
		return sample.Percentile(0.95)
	case "pct99":
		return sample.Percentile(0.99)
	case "pct999":
		return sample.Percentile(0.999)
	}
	return 0
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
