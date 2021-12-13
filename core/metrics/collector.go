package metrics

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	gm "github.com/rcrowley/go-metrics"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

var (
	metricsCfg       *config
	metricsCollector *collector
)

type collector struct {
	collectors map[metricsType]map[string]*metricValue
	locks      map[metricsType]*sync.RWMutex
	httpCli    *fasthttp.Client
}

type request struct {
	Metric    string            `json:"metric"`
	Tags      map[string]string `json:"tags"`
	Value     float64           `json:"value"`
	Timestamp uint64            `json:"timestamp"`
}

type config struct {
	enableMetrics bool
	domain        string
	prefix        string
	printLog      bool //whether print logs during collecting metrics
	flushInterval time.Duration
}

type metricValue struct {
	value   interface{}
	updated bool // When there is a new report, updated is true, otherwise it is false
	lock    *sync.RWMutex
}

type timerValue struct {
	sample gm.Sample
	count  int64
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
	startReport()
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
	getOrStoreDefaultMetric(metricsTypeStore, collectKey, &metricValue{
		value:   float64(0),
		updated: false,
		lock:    &sync.RWMutex{},
	})
	updateMetric(metricsTypeStore, collectKey, value)
}

// 统计本次上报期间（flushInterval）内(name,tags)对应value的累加值，每次上报完需清空
func emitCounter(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	getOrStoreDefaultMetric(metricsTypeCounter, collectKey, &metricValue{
		value:   float64(0),
		updated: false,
		lock:    &sync.RWMutex{},
	})
	updateMetric(metricsTypeCounter, collectKey, value)
}

func emitTimer(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	getOrStoreDefaultMetric(metricsTypeTimer, collectKey, &metricValue{
		value: &timerValue{
			sample: gm.NewUniformSample(reservoirSize),
			count:  0,
		},
		updated: false,
		lock:    &sync.RWMutex{},
	})
	updateMetric(metricsTypeTimer, collectKey, value)
}

func getOrStoreDefaultMetric(metricType metricsType, collectKey string, defaultValue *metricValue) {
	if metricsCollector.collectors[metricType][collectKey] == nil {
		metricsCollector.locks[metricType].Lock()
		defer metricsCollector.locks[metricType].Unlock()
		if metricsCollector.collectors[metricType][collectKey] == nil {
			metricsCollector.collectors[metricType][collectKey] = defaultValue
		}
	}
}

func updateMetric(metricType metricsType, collectKey string, value float64) {
	metricsCollector.locks[metricType].RLock()
	defer metricsCollector.locks[metricType].RUnlock()
	oldValue := metricsCollector.collectors[metricType][collectKey]
	oldValue.lock.Lock()
	defer oldValue.lock.Unlock()
	oldValue.updated = true
	switch metricType {
	case metricsTypeStore:
		oldValue.value = value
	case metricsTypeCounter:
		oldValue.value = oldValue.value.(float64) + value
	case metricsTypeTimer:
		oldValue.value.(*timerValue).sample.Update(int64(value))
		oldValue.value.(*timerValue).count++
	}
}

func startReport() {
	if !isEnableMetrics() {
		return
	}
	flushStore()
	flushCounter()
	flushTimer()
}

func flushStore() {
	ticker := time.NewTicker(metricsCfg.flushInterval)
	go func() {
		for {
			<-ticker.C
			metricsRequests := make([]*Metric, 0)
			metricsCollector.locks[metricsTypeStore].RLock()
			for key, metric := range metricsCollector.collectors[metricsTypeStore] {
				name, tagKvs, ok := parseNameAndTags(key)
				if !ok {
					continue
				}
				metric.lock.Lock()
				if metric.updated { // if updated is false, means no metric emit
					metricsRequest := &Metric{
						Metric:    metricsCfg.prefix + "." + name,
						Tags:      tagKvs,
						Value:     metric.value.(float64),
						Timestamp: uint64(time.Now().Unix()),
					}
					metricsRequests = append(metricsRequests, metricsRequest)
					// reset updated tag after report
					metric.updated = false
					metric.value = float64(0)
				}
				metric.lock.Unlock()
			}
			metricsCollector.locks[metricsTypeStore].RUnlock()
			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("[Metrics] exec store fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("[Metrics] exec store success, url:%s, metricsRequests:%+v", url, metricsRequests)
				}
			}
		}
	}()
}

func flushCounter() {
	ticker := time.NewTicker(metricsCfg.flushInterval)
	go func() {
		for {
			<-ticker.C
			metricsRequests := make([]*Metric, 0)
			metricsCollector.locks[metricsTypeCounter].RLock()
			for key, metric := range metricsCollector.collectors[metricsTypeCounter] {
				name, tagKvs, ok := parseNameAndTags(key)
				if !ok {
					continue
				}
				metric.lock.Lock()
				if metric.updated {
					metricsRequest := &Metric{
						Metric:    metricsCfg.prefix + "." + name,
						Tags:      tagKvs,
						Value:     metric.value.(float64),
						Timestamp: uint64(time.Now().Unix()),
					}
					metricsRequests = append(metricsRequests, metricsRequest)
					// reset updated tag after report
					metric.updated = false
					// After each flushInterval of the counter is reported, the accumulated metric needs to be cleared
					metric.value = float64(0)
				}
				metric.lock.Unlock()
			}
			metricsCollector.locks[metricsTypeCounter].RUnlock()

			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(counterUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("[Metrics] exec counter fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("[Metrics] exec counter success, url:%s, metricsRequests:%+v", url, metricsRequests)
				}
			}
		}
	}()
}

func flushTimer() {
	ticker := time.NewTicker(metricsCfg.flushInterval)
	go func() {
		for {
			<-ticker.C
			metricsRequests := make([]*Metric, 0)
			metricsCollector.locks[metricsTypeTimer].RLock()
			for key, metric := range metricsCollector.collectors[metricsTypeTimer] {
				name, tagKvs, ok := parseNameAndTags(key)
				if !ok {
					return
				}
				metric.lock.Lock()
				if metric.updated {
					timestamp := time.Now().Unix()
					snapshot := metric.value.(*timerValue).sample.Snapshot()
					count := metric.value.(*timerValue).count

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
					metric.value.(*timerValue).count = 0
				}
				metric.lock.Unlock()
			}
			metricsCollector.locks[metricsTypeTimer].RUnlock()
			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("[Metrics] exec timer fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("[Metrics] exec timer success, url:%s, metricsRequests:%+v", url, metricsRequests)
				}
			}
		}
	}()
}

// send send httpRequest to metrics server
func send(metricRequests *MetricMessage, url string) bool {
	for i := 0; i < maxTryTimes; i++ {
		request, err := buildMetricsRequest(metricRequests, url)
		if err != nil {
			fasthttp.ReleaseRequest(request)
			continue
		}
		if doSend(request) {
			return true
		}
	}
	return false
}

func doSend(request *fasthttp.Request) bool {
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	err := metricsCollector.httpCli.DoTimeout(request, response, defaultHttpTimeout)
	if err == nil && response.StatusCode() == fasthttp.StatusOK {
		return true
	}
	if isEnablePrintLog() {
		logs.Error("[Metrics] do http request occur error:%+v\n response:\n%+v", err, response)
	}
	return false
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
