package metrics

import (
	"fmt"
	"strings"

	"github.com/byteplus-sdk/byteplus-sdk-go-rec-core/metrics/protocol"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/proto"
)

type reporter struct {
	httpCli    *fasthttp.Client
	metricsCfg *Config
}

func (r *reporter) reportMetrics(metricMessage *protocol.MetricMessage, url string) error {
	reqBytes, err := proto.Marshal(metricMessage)
	if err != nil {
		return fmt.Errorf("[MetricsReporter] marshal request fail, err:%v url:%s", err, url)
	}
	headers := buildMetricsHeaders()
	return r.doRequest(url, reqBytes, headers)
}

func (r *reporter) reportMetricsLog(metricLogMessage *protocol.MetricLogMessage, url string) error {
	reqBytes, err := proto.Marshal(metricLogMessage)
	if err != nil {
		return fmt.Errorf("[MetricsReporter] marshal request fail, err:%v url:%s", err, url)
	}
	headers := buildMetricsHeaders()
	return r.doRequest(url, reqBytes, headers)
}

func (r *reporter) doRequest(url string, reqBytes []byte, headers map[string]string) error {
	var err error
	request := fasthttp.AcquireRequest()
	response := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseRequest(request)
		fasthttp.ReleaseResponse(response)
	}()
	request.Header.SetMethod(fasthttp.MethodPost)
	request.SetRequestURI(url)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	request.SetBodyRaw(reqBytes)

	for i := 0; i < maxTryTimes; i++ {
		err = r.httpCli.DoTimeout(request, response, r.metricsCfg.HTTPTimeout)
		if err == nil {
			if response.StatusCode() == fasthttp.StatusOK {
				return nil
			}
			return fmt.Errorf("do http request fail, code:%d, rsp: %s",
				response.StatusCode(), response.Body())
		}
		// retry when http timeout
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			err = fmt.Errorf("request timeout, msg: %+v", err)
			continue
		}
		// when occur other err, return directly
		return err
	}
	return err
}

func buildMetricsHeaders() map[string]string {
	headers := make(map[string]string, 2)
	headers["Content-Type"] = "application/x-protobuf"
	headers["Accept"] = "application/json"
	return headers
}
