package metrics

import (
	"fmt"
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"time"
)

type Store struct {
	name       string
	expireTime time.Time
	queue      chan *Item
	valueMap   map[Item]*MetricsRequest
	httpCli    *Client
}

func NewStore(name string) *Store {
	return NewStoreWithFlushTime(name, defaultFlushInterval)
}

func NewStoreWithFlushTime(name string, flushInterval time.Duration) *Store {
	c := &Store{
		name:       name,
		expireTime: time.Now().Add(defaultMetricsExpireTime),
		queue:      make(chan *Item, maxQueueSize),
		valueMap:   make(map[Item]*MetricsRequest),
		httpCli:    GetClient(fmt.Sprintf(otherUrlFormat, metricsDomain)),
	}
	return c
}

func (c *Store) emit(value float64, tags map[string]string) {
	tag := processTags(tags)
	item := NewItem(tag, value)
	select {
	case c.queue <- item:
	default:
		if IsEnablePrintLog() {
			logs.Warn("metrics emit too fast, exceed max queue size(%d)", maxQueueSize)
		}
	}
}

func (c *Store) isExpired() bool {
	return time.Now().After(c.expireTime)
}

func (c *Store) updateExpireTime(ttl time.Duration) {
	if ttl > 0 {
		c.expireTime = time.Now().Add(ttl)
	}
}

func (c *Store) getName() string {
	return c.name
}

func (c *Store) flush() {
	defer func() {
		if err := recover(); err != nil {
			logs.Error("exec store err: %v", err)
		}
	}()
	for size := 0; size < maxFlashSize && len(c.queue) != 0; size++ {
		item := <-c.queue
		if req, ok := c.valueMap[*item]; ok {
			req.Value = item.value
		} else {
			metricsRequest := &MetricsRequest{
				MetricsName: c.name,
				Value:       item.value,
				TagKvs:      recoverTags(item.tags),
			}
			c.valueMap[*item] = metricsRequest
		}
	}
	metricsRequests := make([]*MetricsRequest, 0, len(c.valueMap))
	if len(c.valueMap) != 0 {
		timestamp := time.Now().Unix()
		for item, metricsRequest := range c.valueMap {
			metricsRequest.TimeStamp = timestamp
			metricsRequests = append(metricsRequests, metricsRequest)
			delete(c.valueMap, item)
			if IsEnablePrintLog() {
				logs.Info("remove counter key: %+v", item)
			}
		}
		c.httpCli.put(metricsRequests)
	}
}
