package core

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/byteplus-sdk/sdk-go/core/metrics"

	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/valyala/fasthttp"
)

const (
	defaultPingURLFormat = "{}://%s/predict/api/ping"
	defaultWindowSize    = 60
	defaultPingTimeout   = 300 * time.Millisecond
	defaultPingInterval  = time.Second
	failureRateThreshold = 0.1
)

type HostAvailablerConfig struct {
	// host availabler used to test the latency, example {}://%s/predict/api/ping
	// {} will be replaced by schema which set in context
	// %s will be dynamically formatted by hosts
	PingUrlFormat string
	// record the window size of each host's test status
	WindowSize int
	// timeout for requesting hosts
	PingTimeout time.Duration
	// The time interval for pingHostAvailabler to do ping
	PingInterval time.Duration
}

func NewHostAvailabler(urlCenter URLCenter, context *Context) *HostAvailabler {
	availabler := &HostAvailabler{
		context:   context,
		urlCenter: urlCenter,
		config:    fillDefaultConfig(context.hostAvailablerConfig),
	}
	availabler.currentHost = context.hosts[0]
	availabler.availableHosts = context.hosts
	availabler.pingUrlFormat = strings.ReplaceAll(availabler.config.PingUrlFormat, "{}", context.Schema())
	if len(context.hosts) <= 1 {
		return availabler
	}
	hostWindowMap := make(map[string]*window, len(context.hosts))
	hostHttpCliMap := make(map[string]*fasthttp.HostClient, len(context.hosts))
	for _, host := range context.hosts {
		hostWindowMap[host] = newWindow(availabler.config.WindowSize)
		hostHttpCliMap[host] = &fasthttp.HostClient{Addr: host}
	}
	availabler.hostWindowMap = hostWindowMap
	availabler.hostHTTPCliMap = hostHttpCliMap
	AsyncExecute(availabler.scheduleFunc())
	return availabler
}

func fillDefaultConfig(config *HostAvailablerConfig) *HostAvailablerConfig {
	if config == nil {
		config = &HostAvailablerConfig{}
	}
	if config.PingUrlFormat == "" {
		config.PingUrlFormat = defaultPingURLFormat
	}
	if config.PingTimeout <= 0 {
		config.PingTimeout = defaultPingTimeout
	}
	if config.WindowSize <= 0 {
		config.WindowSize = defaultWindowSize
	}
	if config.PingInterval <= 0 {
		config.PingInterval = defaultPingInterval
	}
	return config
}

type HostAvailabler struct {
	abort          bool
	context        *Context
	urlCenter      URLCenter
	config         *HostAvailablerConfig
	currentHost    string
	availableHosts []string
	hostWindowMap  map[string]*window
	hostHTTPCliMap map[string]*fasthttp.HostClient
	pingUrlFormat  string
}

func (receiver *HostAvailabler) Shutdown() {
	receiver.abort = true
}

func (receiver *HostAvailabler) scheduleFunc() func() {
	return func() {
		ticker := time.NewTicker(receiver.config.PingInterval)
		for true {
			if receiver.abort {
				ticker.Stop()
				return
			}
			receiver.checkHost()
			receiver.switchHost()
			<-ticker.C
		}
	}
}

func (receiver *HostAvailabler) checkHost() {
	availableHosts := make([]string, 0, len(receiver.context.hosts))
	for _, host := range receiver.context.hosts {
		winObj := receiver.hostWindowMap[host]
		winObj.put(receiver.ping(host))
		if winObj.failureRate() < failureRateThreshold {
			availableHosts = append(availableHosts, host)
		}
	}
	receiver.availableHosts = availableHosts
	if len(availableHosts) <= 1 {
		return
	}
	sort.Slice(availableHosts, func(i, j int) bool {
		failureRateI := receiver.hostWindowMap[availableHosts[i]].failureRate()
		failureRateJ := receiver.hostWindowMap[availableHosts[j]].failureRate()
		return failureRateI < failureRateJ
	})
}

func (receiver *HostAvailabler) ping(host string) bool {
	start := time.Now()
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	url := fmt.Sprintf(receiver.pingUrlFormat, host)
	request.SetRequestURI(url)
	request.Header.SetMethod(fasthttp.MethodGet)
	for k, v := range receiver.context.CustomerHeaders() {
		request.Header.Set(k, v)
	}
	reqID := uuid.NewString()
	request.Header.Set("Request-Id", reqID)
	request.Header.Set("Tenant", receiver.context.Tenant())
	if len(receiver.context.hostHeader) > 0 {
		request.SetHost(receiver.context.hostHeader)
	}
	httpCli := receiver.hostHTTPCliMap[host]
	err := httpCli.DoTimeout(request, response, receiver.config.PingTimeout)
	cost := time.Now().Sub(start)
	if err != nil {
		metrics.Warn(reqID, "[ByteplusSDK] ping find err, tenant:%s, host:%s, cost:%dms, err:%v",
			receiver.context.Tenant(), host, cost.Milliseconds(), err)
		logs.Warn("ping find err, host:%s cost:%dms err:%v", host, cost.Milliseconds(), err)
	}
	if response.StatusCode() == fasthttp.StatusOK {
		metrics.Info(reqID, "[ByteplusSDK] ping success, tenant:%s, host:%s, cost:%dms",
			receiver.context.Tenant(), host, cost.Milliseconds())
		logs.Debug("ping success host:'%s' cost:'%s'", host, cost)
		return true
	}
	metrics.Warn(reqID, "[ByteplusSDK] ping fail, tenant:%s, host:%s, cost:%dms, status:%d",
		receiver.context.Tenant(), host, cost.Milliseconds(), response.StatusCode())
	logs.Warn("ping fail, host:%s cost:%s status:%d err:%v",
		host, cost, response.StatusCode(), err)
	return false
}

func (receiver *HostAvailabler) switchHost() {
	var newHost string
	if len(receiver.availableHosts) == 0 {
		newHost = receiver.context.hosts[0]
	} else {
		newHost = receiver.availableHosts[0]
	}
	if newHost != receiver.currentHost {
		logs.Warn("switch host to '%s', origin is '%s'",
			newHost, receiver.currentHost)
		receiver.currentHost = newHost
		receiver.urlCenter.Refresh(newHost)
		if receiver.context.hostHeader != "" {
			receiver.context.hostHTTPCli = &fasthttp.HostClient{Addr: newHost}
		}
	}
}

func (receiver *HostAvailabler) GetHost() string {
	return receiver.currentHost
}

func newWindow(size int) *window {
	result := &window{
		size:         size,
		items:        make([]bool, size),
		head:         size - 1,
		tail:         0,
		failureCount: 0,
	}
	for i := range result.items {
		result.items[i] = true
	}
	return result
}

type window struct {
	size         int
	items        []bool
	head         int
	tail         int
	failureCount float64
}

func (receiver *window) put(success bool) {
	if !success {
		receiver.failureCount++
	}
	receiver.head = (receiver.head + 1) % receiver.size
	receiver.items[receiver.head] = success
	receiver.tail = (receiver.tail + 1) % receiver.size
	removingItem := receiver.items[receiver.tail]
	if !removingItem {
		receiver.failureCount--
	}
}

func (receiver *window) failureRate() float64 {
	return receiver.failureCount / float64(receiver.size)
}

func (receiver *window) String() string {
	return fmt.Sprintf("%+v", *receiver)
}
