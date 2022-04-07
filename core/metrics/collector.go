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
	timerStatMetrics = []string{"max", "min", "avg", "pct75", "pct90", "pct95", "pct99", "pct999"}
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
	httpTimeoutMs time.Duration
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
		httpTimeoutMs: defaultHttpTimeout,
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
		if flushInterval > 5000*time.Millisecond { // flushInterval should not be too small
			config.flushInterval = flushInterval
		}
	}
}

// WithMetricsTimeout set the interval of reporting metrics
func WithMetricsTimeout(timeout time.Duration) Option {
	return func(config *config) {
		if timeout > defaultHttpTimeout {
			config.httpTimeoutMs = timeout
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

// Update the value corresponding to (name, tags) to the latest value,
// and there is no need to clear it after each report.
func emitStore(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	updateMetric(metricsTypeStore, collectKey, value)
}

// Count the accumulated value of (name, tags) corresponding to the
// value during this reporting period (flushInterval), and it needs
// to be cleared after each reporting period
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
	metric := getOrCreateMetric(metricType, collectKey)
	switch metricType {
	case metricsTypeStore:
		metric.value = value
	case metricsTypeCounter:
		metric.value.(*atomic.Float64).Add(value)
	case metricsTypeTimer:
		metric.value.(Sample).Update(int64(value))
	}
	metric.updated = true
}

func getOrCreateMetric(metricType metricsType, collectKey string) *metricValue {
	metricsCollector.locks[metricType].RLock()
	if metricsCollector.collectors[metricType][collectKey] != nil {
		metricsCollector.locks[metricType].RUnlock()
		return metricsCollector.collectors[metricType][collectKey]
	}
	metricsCollector.locks[metricType].RUnlock()

	// set default metric
	metricsCollector.locks[metricType].Lock()
	defer metricsCollector.locks[metricType].Unlock()
	if metricsCollector.collectors[metricType][collectKey] == nil {
		metricsCollector.collectors[metricType][collectKey] = buildDefaultMetric(metricType)
	}
	return metricsCollector.collectors[metricType][collectKey]
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
			// reset updated tag after report
			metric.updated = false
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
		}
	}
	metricsCollector.locks[metricsTypeStore].RUnlock()
	if len(metricsRequests) > 0 {
		url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
		sendMetrics(metricsRequests, url)
	}
}

func flushCounter() {
	metricsRequests := make([]*Metric, 0, len(metricsCollector.collectors[metricsTypeCounter]))
	metricsCollector.locks[metricsTypeCounter].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeCounter] {
		if metric.updated {
			// reset updated tag after report
			metric.updated = false
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
		sendMetrics(metricsRequests, url)
	}
}

func flushTimer() {
	metricsRequests := make([]*Metric, 0, len(metricsCollector.collectors[metricsTypeTimer])*len(timerStatMetrics))
	metricsCollector.locks[metricsTypeTimer].RLock()
	for key, metric := range metricsCollector.collectors[metricsTypeTimer] {
		if metric.updated {
			// reset updated tag after report
			metric.updated = false
			name, tagKvs, ok := parseNameAndTags(key)
			if !ok {
				return
			}
			snapshot := metric.value.(Sample).Snapshot()
			// clear sample every sample period
			metric.value.(Sample).Clear()
			metricsRequests = append(metricsRequests, buildStatMetrics(snapshot, name, tagKvs)...)
		}
	}
	metricsCollector.locks[metricsTypeTimer].RUnlock()
	if len(metricsRequests) > 0 {
		url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
		sendMetrics(metricsRequests, url)
	}
}

func sendMetrics(metricsRequests []*Metric, url string) {
	if err := send(&MetricMessage{Metrics: metricsRequests}, url); err != nil {
		logs.Error("[BytePlusSDK][Metrics] send metrics err:%+v, url:%s, metricsRequests:%+v", err, url, metricsRequests)
		return
	}
	if isEnablePrintLog() {
		logs.Debug("[BytePlusSDK][Metrics] send metrics success, url:%s, metricsRequests:%+v", url, metricsRequests)
	}
}

func buildStatMetrics(sample Sample, name string, tagKvs map[string]string) []*Metric {
	timestamp := uint64(time.Now().Unix())
	metricsRequests := make([]*Metric, 0, len(timerStatMetrics))
	for _, statName := range timerStatMetrics {
		value, ok := getStatValue(statName, sample)
		if !ok {
			continue
		}
		metricsRequests = append(metricsRequests, &Metric{
			Metric:    metricsCfg.prefix + "." + name + "." + statName,
			Tags:      tagKvs,
			Value:     value,
			Timestamp: timestamp,
		})
	}
	return metricsRequests
}

func getStatValue(statName string, sample Sample) (float64, bool) {
	switch statName {
	case "max":
		return float64(sample.Max()), true
	case "min":
		return float64(sample.Min()), true
	case "avg":
		return sample.Mean(), true
	case "pct75":
		return sample.Percentile(0.75), true
	case "pct90":
		return sample.Percentile(0.90), true
	case "pct95":
		return sample.Percentile(0.95), true
	case "pct99":
		return sample.Percentile(0.99), true
	case "pct999":
		return sample.Percentile(0.999), true
	}
	return 0, false
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
	err := metricsCollector.httpCli.DoTimeout(request, response, metricsCfg.httpTimeoutMs)
	if err == nil && response.StatusCode() == fasthttp.StatusOK {
		return nil
	}
	return err
}
