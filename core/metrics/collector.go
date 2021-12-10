package metrics

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	gm "github.com/rcrowley/go-metrics"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
	"time"
)

var (
	metricsCfg       *config
	metricsCollector *collector
)

type collector struct {
	collectors map[metricsType]map[string]*atomic.Value
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

type timer struct {
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
		collectors: map[metricsType]map[string]*atomic.Value{
			metricsTypeCounter: make(map[string]*atomic.Value),
			metricsTypeTimer:   make(map[string]*atomic.Value),
			metricsTypeStore:   make(map[string]*atomic.Value),
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
	if metricsCollector.collectors[metricsTypeStore][collectKey] == nil {
		metricsCollector.locks[metricsTypeStore].Lock()
		if metricsCollector.collectors[metricsTypeStore][collectKey] == nil {
			metricsCollector.collectors[metricsTypeStore][collectKey] = &atomic.Value{}
			metricsCollector.collectors[metricsTypeStore][collectKey].Store(float64(0))
		}
		metricsCollector.locks[metricsTypeStore].Unlock()
	}
	metricsCollector.locks[metricsTypeStore].RLock()
	defer metricsCollector.locks[metricsTypeStore].RUnlock()
	metricsCollector.collectors[metricsTypeStore][collectKey].Store(value)
}

// 统计本次上报期间（flushInterval）内(name,tags)对应value的累加值，每次上报完需清空
func emitCounter(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	if metricsCollector.collectors[metricsTypeCounter][collectKey] == nil {
		metricsCollector.locks[metricsTypeCounter].Lock()
		if metricsCollector.collectors[metricsTypeCounter][collectKey] == nil {
			metricsCollector.collectors[metricsTypeCounter][collectKey] = &atomic.Value{}
			metricsCollector.collectors[metricsTypeCounter][collectKey].Store(float64(0))
		}
		metricsCollector.locks[metricsTypeCounter].Unlock()
	}
	metricsCollector.locks[metricsTypeCounter].RLock()
	defer metricsCollector.locks[metricsTypeCounter].RUnlock()
	oldValue := metricsCollector.collectors[metricsTypeCounter][collectKey].Load().(float64)
	metricsCollector.collectors[metricsTypeCounter][collectKey].Store(oldValue + value)
}

func emitTimer(name string, value float64, tagKvs ...string) {
	if !isEnableMetrics() {
		return
	}
	collectKey := buildCollectKey(name, tagKvs)
	if metricsCollector.collectors[metricsTypeTimer][collectKey] == nil {
		metricsCollector.locks[metricsTypeTimer].Lock()
		if metricsCollector.collectors[metricsTypeTimer][collectKey] == nil {
			metricsCollector.collectors[metricsTypeTimer][collectKey] = &atomic.Value{}
			metricsCollector.collectors[metricsTypeTimer][collectKey].Store(&timer{
				sample: gm.NewUniformSample(reservoirSize),
				count:  0,
			})
		}
		metricsCollector.locks[metricsTypeTimer].Unlock()
	}
	metricsCollector.locks[metricsTypeTimer].RLock()
	defer metricsCollector.locks[metricsTypeTimer].RUnlock()
	oldValue := metricsCollector.collectors[metricsTypeTimer][collectKey].Load().(*timer)
	oldValue.sample.Update(int64(value))
	atomic.AddInt64(&oldValue.count, 1)
	metricsCollector.collectors[metricsTypeTimer][collectKey].Store(oldValue)
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
			for key, value := range metricsCollector.collectors[metricsTypeStore] {
				name, tagKvs, ok := parseNameAndTags(key)
				if ok {
					metricsRequest := &Metric{
						Metric:    metricsCfg.prefix + "." + name,
						Tags:      tagKvs,
						Value:     value.Load().(float64),
						Timestamp: uint64(time.Now().Unix()),
					}
					metricsRequests = append(metricsRequests, metricsRequest)
				}
			}
			metricsCollector.locks[metricsTypeStore].RUnlock()
			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("exec store fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("exec store success, url:%s, metricsRequests:%+v", url, metricsRequests)
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
			for key, value := range metricsCollector.collectors[metricsTypeCounter] {
				name, tagKvs, ok := parseNameAndTags(key)
				if ok {
					metricsRequest := &Metric{
						Metric:    metricsCfg.prefix + "." + name,
						Tags:      tagKvs,
						Value:     value.Load().(float64),
						Timestamp: uint64(time.Now().Unix()),
					}
					metricsRequests = append(metricsRequests, metricsRequest)
				}
				// After each flushInterval of the counter is reported, the accumulated value needs to be cleared
				value.Store(float64(0))
			}
			metricsCollector.locks[metricsTypeCounter].RUnlock()

			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(counterUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("exec counter fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("exec counter success, url:%s, metricsRequests:%+v", url, metricsRequests)
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
			for key, value := range metricsCollector.collectors[metricsTypeTimer] {
				name, tagKvs, ok := parseNameAndTags(key)
				if ok {
					timestamp := time.Now().Unix()
					snapshot := value.Load().(*timer).sample.Snapshot()
					count := atomic.LoadInt64(&value.Load().(*timer).count)

					//qps
					//The timer data will be transmitted to metrics_proxy and reported as store value,
					//so the qps here must also be store value, which need to be divided by flushInterval,
					//and the qps value should be showed as store in grafana.
					metricsRequests = append(metricsRequests, &Metric{
						Metric:    metricsCfg.prefix + "." + name + "." + "qps",
						Tags:      tagKvs,
						Value:     float64(count) / metricsCfg.flushInterval.Seconds(),
						Timestamp: uint64(timestamp),
					})
					//count is the counter value and needs to be cleared after each report
					atomic.StoreInt64(&value.Load().(*timer).count, 0)
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
					//pc990
					metricsRequests = append(metricsRequests, &Metric{
						Metric:    metricsCfg.prefix + "." + name + "." + "pct999",
						Tags:      tagKvs,
						Value:     snapshot.Percentile(0.999),
						Timestamp: uint64(timestamp),
					})
				}
			}
			metricsCollector.locks[metricsTypeTimer].RUnlock()
			if len(metricsRequests) > 0 {
				url := fmt.Sprintf(otherUrlFormat, metricsCfg.domain)
				if !send(&MetricMessage{Metrics: metricsRequests}, url) {
					if isEnablePrintLog() {
						logs.Error("exec timer fail, url:%s", url)
					}
					continue
				}
				if isEnablePrintLog() {
					logs.Debug("exec timer success, url:%s, metricsRequests:%+v", url, metricsRequests)
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
		logs.Error("do http request occur error:%+v\n response:\n%+v", err, response)
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
