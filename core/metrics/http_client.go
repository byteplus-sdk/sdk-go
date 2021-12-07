package metrics

import (
	"encoding/json"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"github.com/valyala/fasthttp"
	"sync"
	"time"
)

const (
	maxTryTimes        = 2
	maxRequestsSize    = 5000 //todo:确认需要多大
	defaultHttpTimeout = 800 * time.Millisecond
)

var (
	httpCli     = &fasthttp.Client{}
	clientCache = &InstanceCache{
		instanceMap:     make(map[string]interface{}),
		instanceBuilder: newClient,
		lock:            &sync.Mutex{},
	}
)

type MetricsRequest struct {
	MetricsName string            `json:"metric"`
	TagKvs      map[string]string `json:"tags"`
	Value       float64           `json:"value"`
	TimeStamp   int64             `json:"timestamp"`
}

type Client struct {
	url      string
	timeout  time.Duration
	requests chan []*MetricsRequest
	stopped  chan struct{}
}

func GetClient(url string) *Client {
	return clientCache.GetInstanceByName(url).(*Client)
}

func newClient(url string) interface{} {
	cli := &Client{
		url:      url,
		timeout:  defaultHttpTimeout,
		requests: make(chan []*MetricsRequest, maxRequestsSize),
		stopped:  make(chan struct{}),
	}
	cli.start()
	return cli
}

func (h *Client) start() {
	AsyncExecute(func() {
		for {
			select {
			case metricsRequests := <-h.requests:
				if success := h.send(metricsRequests); !success {
					logs.Error("exec metrics fail, url:%s", h.url)
				}
			case <-h.stopped:
				return
			}
		}
	})
}

// todo:如何关闭单例中的任务，可以采用定期回收的方式
func (h *Client) stop() {
	h.stopped <- struct{}{}
}

func (h *Client) put(metricRequests []*MetricsRequest) {
	select {
	case h.requests <- metricRequests:
	default:
		if IsEnablePrintLog() {
			logs.Warn("metrics requests emit too fast, exceed max queue size(%d)", maxRequestsSize)
		}
	}
}

// send send httpRequest to metrics server
func (h *Client) send(metricRequests []*MetricsRequest) bool {
	for i := 0; i < maxTryTimes; i++ {
		request, err := h.buildMetricsRequest(metricRequests)
		if err != nil {
			fasthttp.ReleaseRequest(request)
			continue
		}
		if h.doSend(request) {
			return true
		}
	}
	return false
}

func (h *Client) emit(metricRequest []*MetricsRequest) bool {
	return h.send(metricRequest)
}

func (h *Client) doSend(request *fasthttp.Request) bool {
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	var err error
	if h.timeout > 0 {
		err = httpCli.DoTimeout(request, response, h.timeout)
	} else {
		err = httpCli.Do(request, response)
	}
	if err == nil && response.StatusCode() == fasthttp.StatusOK {
		if IsEnablePrintLog() {
			logs.Debug("success reporting metrics request:\n%+v", request)
		}
		return true
	}
	if IsEnablePrintLog() {
		logs.Error("do http request occur error:%+v\n url:%s\n response:\n%+v", err, h.url, response)
	}
	return false
}

func (h *Client) buildMetricsRequest(metricRequests []*MetricsRequest) (*fasthttp.Request, error) {
	request := fasthttp.AcquireRequest()
	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(h.url)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	body, err := json.Marshal(metricRequests)
	if err != nil {
		return nil, err
	}
	request.SetBodyRaw(body)
	return request, nil
}
