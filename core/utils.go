package core

import (
	"github.com/byteplus-sdk/sdk-go/core/metrics"
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func AsyncExecute(runnable func()) {
	go func(run func()) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[ByteplusSDK] async execute occur panic, "+
					"please feedback to bytedance, err:%v trace:\n%s", r, string(debug.Stack()))
			}
		}()
		run()
	}(runnable)
}

func IsNetError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), netErrMark)
}

func IsTimeoutError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "timeout")
}

func ReportRequestSuccess(metricsPrefix, url string, begin time.Time) {
	urlTag := buildUrlTags(url)
	metrics.Latency(buildLatencyKey(metricsPrefix), begin, urlTag...)
	metrics.Counter(buildCounterKey(metricsPrefix), 1, urlTag...)
}

func ReportRequestError(metricsPrefix, url string, begin time.Time, code int, message string) {
	urlTag := buildUrlTags(url)
	tagKvs := append(urlTag, "code:"+strconv.Itoa(code), "message:"+message)
	metrics.Latency(buildLatencyKey(metricsPrefix), begin, tagKvs...)
	metrics.Counter(buildCounterKey(metricsPrefix), 1, tagKvs...)
}

func ReportRequestException(metricsPrefix, url string, begin time.Time, err error) {
	urlTag := buildUrlTags(url)
	withErrorTags(urlTag, err)
	metrics.Latency(buildLatencyKey(metricsPrefix), begin, urlTag...)
	metrics.Counter(buildCounterKey(metricsPrefix), 1, urlTag...)
}

func withErrorTags(urlTag []string, err error) {
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "time") && strings.Contains(msg, "out") {
		if strings.Contains(msg, "connect") {
			urlTag = append(urlTag, "message:connect-timeout")
			return
		}
		if strings.Contains(msg, "read") {
			urlTag = append(urlTag, "message:read-timeout")
			return
		}
		urlTag = append(urlTag, "message:timeout")
		return
	}
	urlTag = append(urlTag, "message:other")
}

func buildUrlTags(url string) []string {
	return []string{"url:" + adjustUrlTag(url), "req_type:" + parseReqType(url)}
}

func adjustUrlTag(url string) string {
	return strings.ReplaceAll(url, "=", "_is_")
}

func parseReqType(url string) string {
	if strings.Contains(url, "ping") {
		return "ping"
	}
	if strings.Contains(url, "data/api") {
		return "data-api"
	}
	if strings.Contains(url, "predict/api") {
		return "predict-api"
	}
	return "unknown"
}

func buildCounterKey(metricsPrefix string) string {
	return metricsPrefix + "." + "count"
}

func buildLatencyKey(metricsPrefix string) string {
	return metricsPrefix + "." + "latency"
}
