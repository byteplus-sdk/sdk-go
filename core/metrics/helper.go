package metrics

import (
	"sort"
	"strings"
	"time"
)

// Counter tagKvs should be formatted as "key:value"
// example: metrics.Counter("request.qps", 1, "method:user", "type:upload")
func Counter(key string, value float64, tagKvs ...string) {
	emitCounter(key, value, tagKvs...)
}

// Timer tagKvs should be formatted as "key:value"
// example: metrics.Timer("request.cost", 100, "method:user", "type:upload")
func Timer(key string, value float64, tagKvs ...string) {
	emitTimer(key, value, tagKvs...)
}

// Latency report time cost for execution
// tagKvs should be formatted as "key:value"
// example: metrics.Latency("request.latency", startTime, "method:user", "type:upload")
func Latency(key string, begin time.Time, tagKvs ...string) {
	emitTimer(key, float64(time.Now().Sub(begin).Milliseconds()), tagKvs...)
}

// Store tagKvs should be formatted as "key:value"
// example: metrics.Store("goroutine.count", 400, "ip:127.0.0.1")
func Store(key string, value float64, tagKvs ...string) {
	emitStore(key, value, tagKvs...)
}

// process tags slice to string in order
func tags2String(tags []string) string {
	sort.Strings(tags)
	return strings.Join(tags, "|")
}

func buildCollectKey(name string, tags []string) string {
	tagString := tags2String(tags)
	return name + delimiter + tagString
}

func parseNameAndTags(src string) (string, map[string]string, bool) {
	index := strings.Index(src, delimiter)
	if index == -1 {
		return "", nil, false
	}
	name := src[:index]
	tagString := src[index+len(delimiter):]
	tagKvs := recoverTags(tagString)
	return name, tagKvs, true
}

// recover tagString to origin Tags map
func recoverTags(tagString string) map[string]string {
	tagKvs := make(map[string]string)
	kvs := strings.Split(tagString, "|")
	for _, kv := range kvs {
		res := strings.SplitN(kv, ":", 2)
		if len(res) < 2 {
			continue
		}
		tagKvs[res[0]] = res[1]
	}
	return tagKvs
}
