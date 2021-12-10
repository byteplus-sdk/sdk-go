package metrics

import (
	"sort"
	"strings"
	"time"
)

func Counter(key string, value float64, tagKvs ...string) {
	emitCounter(key, value, tagKvs...)
}

/**
 * Timer介绍：https://site.bytedance.net/docs/2080/2717/36907/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Timer("request.cost", 100, "method:user")
 */

func Timer(key string, value float64, tagKvs ...string) {
	emitTimer(key, value, tagKvs...)
}

/**
 * Latency介绍：Latency基于timer封装，非Metrics的标准类型。timer介绍：https://site.bytedance.net/docs/2080/2717/36907/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Latency("request.cost", startTime, "method:user")
 */

func Latency(key string, begin time.Time, tagKvs ...string) {
	emitTimer(key, float64(time.Now().Sub(begin).Milliseconds()), tagKvs...)
}

/**
 * Store介绍：https://site.bytedance.net/docs/2080/2717/36905/
 * tags：tag列表，格式"key:Value"，通过“:”分割key和value
 * 示例：metrics.Store("goroutine.count", 400, "ip:127.0.0.1")
 */

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
		res := strings.Split(kv, ":")
		if len(res) != 2 {
			continue
		}
		tagKvs[res[0]] = res[1]
	}
	return tagKvs
}
