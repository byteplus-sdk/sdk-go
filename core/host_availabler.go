package core

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/valyala/fasthttp"
	"sort"
	"strings"
	"time"
)

const (
	pingUrlFormat        = "{}://%s/predict/api/ping"
	pingInterval         = time.Second
	windowSize           = 60
	failureRateThreshold = 0.1
	pingTimeout          = 200 * time.Millisecond
)

func NewHostAvailabler(urlCenter URLCenter, context *Context) *HostAvailabler {
	availabler := &HostAvailabler{
		context:   context,
		urlCenter: urlCenter,
	}
	availabler.pingUrlFormat = strings.ReplaceAll(pingUrlFormat, "{}", context.Schema())
	if len(context.hosts) <= 1 {
		return availabler
	}
	availabler.currentHost = context.hosts[0]
	availabler.availableHosts = context.hosts
	hostWindowMap := make(map[string]*window, len(context.hosts))
	hostHttpCliMap := make(map[string]*fasthttp.HostClient, len(context.hosts))
	for _, host := range context.hosts {
		hostWindowMap[host] = newWindow(windowSize)
		hostHttpCliMap[host] = &fasthttp.HostClient{Addr: host}
	}
	availabler.hostWindowMap = hostWindowMap
	availabler.hostHttpCliMap = hostHttpCliMap
	AsyncExecute(availabler.startSchedule())
	return availabler
}

type HostAvailabler struct {
	context        *Context
	urlCenter      URLCenter
	currentHost    string
	availableHosts []string
	hostWindowMap  map[string]*window
	hostHttpCliMap map[string]*fasthttp.HostClient
	pingUrlFormat  string
}

func (receiver *HostAvailabler) startSchedule() func() {
	return func() {
		tick := time.Tick(pingInterval)
		for true {
			receiver.checkHost()
			receiver.switchHost()
			<-tick
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
	if len(receiver.context.hostHeader) > 0 {
		request.SetHost(receiver.context.hostHeader)
	}
	httpCli := receiver.hostHttpCliMap[host]
	err := httpCli.DoTimeout(request, response, pingTimeout)
	cost := time.Now().Sub(start)
	if err == nil && response.StatusCode() == fasthttp.StatusOK {
		logs.Trace("ping success host:'%s' cost:'%s'", host, cost)
		return true
	}
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
		receiver.context.httpCli = &fasthttp.HostClient{Addr: newHost}
	}
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
