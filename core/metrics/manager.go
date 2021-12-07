package metrics

import (
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"sync"
	"sync/atomic"
	"time"
)

var managerCache = &InstanceCache{
	instanceMap:     make(map[string]interface{}),
	instanceBuilder: newManager,
	lock:            &sync.Mutex{},
}

type Manager struct {
	prefix        string
	tidyInterval  time.Duration
	locks         map[metricsType]*sync.RWMutex
	flushInterval time.Duration
	metricsMaps   map[metricsType]map[string]Metrics
	stopped       int32
}

// GetManager return instance of client according metrics prefix
func GetManager(prefix string) *Manager {
	return managerCache.GetInstanceByName(prefix).(*Manager)
}

func newManager(prefix string) interface{} {
	return newManagerWithParams(prefix, defaultTidyInterval, defaultFlushInterval)
}

//tidyInterval should be longer than flushInterval
func newManagerWithParams(prefix string, ttl time.Duration, flushInterval time.Duration) *Manager {
	manager := &Manager{
		prefix:       prefix,
		tidyInterval: ttl,
		locks: map[metricsType]*sync.RWMutex{
			metricsTypeStore:   &sync.RWMutex{},
			metricsTypeCounter: &sync.RWMutex{},
			metricsTypeTimer:   &sync.RWMutex{},
		},
		flushInterval: flushInterval,
		metricsMaps: map[metricsType]map[string]Metrics{
			metricsTypeStore:   make(map[string]Metrics, 256),
			metricsTypeCounter: make(map[string]Metrics, 256),
			metricsTypeTimer:   make(map[string]Metrics, 256),
		},
		stopped: 0,
	}
	manager.start()
	return manager
}

func (m *Manager) start() {
	atomic.StoreInt32(&m.stopped, 0)
	// Regularly report metrics, one thread for each type
	for metricsType := range m.metricsMaps {
		metricsType := metricsType
		go func() {
			ticker := time.NewTicker(m.flushInterval)
			for {
				if m.isStopped() {
					return
				}
				<-ticker.C
				m.reportMetrics(metricsType)
			}
		}()
	}
	// Regularly clean up overdue metrics
	go func() {
		ticker := time.NewTicker(m.tidyInterval)
		for {
			if m.isStopped() {
				return
			}
			<-ticker.C
			m.tidy()
		}
	}()
}

func (m *Manager) isStopped() bool {
	return atomic.LoadInt32(&m.stopped) == 1
}

func (m *Manager) Stop() {
	m.tidy()
	atomic.StoreInt32(&m.stopped, 1)
}

func (m *Manager) reportMetrics(metricsType metricsType) {
	m.locks[metricsType].RLock()
	defer m.locks[metricsType].RUnlock()
	for _, metrics := range m.metricsMaps[metricsType] { //read map
		// report stores
		metrics.flush()
	}
}

func (m *Manager) tidy() {
	// clean expired metrics
	for metricsType, metricsMap := range m.metricsMaps {
		expiredMetrics := make([]Metrics, 0)
		for _, metrics := range metricsMap {
			if metrics.isExpired() {
				expiredMetrics = append(expiredMetrics, metrics)
			}
		}
		if len(expiredMetrics) != 0 {
			m.locks[metricsType].Lock()
			for _, metrics := range expiredMetrics {
				delete(metricsMap, metrics.getName())
				if IsEnablePrintLog() {
					logs.Info("remove expired metrics %+v", metrics.getName())
				}
			}
			m.locks[metricsType].Unlock()
		}
	}
}

func (m *Manager) emitCounter(name string, tags map[string]string, value float64) {
	m.getOrAddMetrics(metricsTypeCounter, m.prefix+"."+name, nil).emit(value, tags)
}

func (m *Manager) emitTimer(name string, tags map[string]string, value float64) {
	m.getOrAddMetrics(metricsTypeTimer, m.prefix+"."+name, tags).emit(value, nil)
}

func (m *Manager) emitStore(name string, tags map[string]string, value float64) {
	m.getOrAddMetrics(metricsTypeStore, m.prefix+"."+name, nil).emit(value, tags)
}

func (m *Manager) getOrAddMetrics(metricsType metricsType, name string, tagKvs map[string]string) Metrics {
	tagString := ""
	if len(tagKvs) != 0 {
		tagString = processTags(tagKvs)
	}
	metricsKey := name + tagString
	metrics, exist := m.metricsMaps[metricsType][metricsKey]
	if !exist {
		m.locks[metricsType].Lock()
		defer m.locks[metricsType].Unlock()
		metrics, exist := m.metricsMaps[metricsType][metricsKey]
		if !exist {
			metrics = m.buildMetrics(metricsType, name, tagString)
			metrics.updateExpireTime(m.tidyInterval)
			m.metricsMaps[metricsType][metricsKey] = metrics
			return metrics
		}
	}
	return metrics
}

func (m *Manager) buildMetrics(metricsType metricsType, name string, tagString string) Metrics {
	switch metricsType {
	case metricsTypeStore:
		return NewStoreWithFlushTime(name, m.flushInterval)
	case metricsTypeCounter:
		return NewCounterWithFlushTime(name, m.flushInterval)
	case metricsTypeTimer:
		return NewTimerWithFlushTime(name, tagString, m.flushInterval)
	}
	return nil
}
