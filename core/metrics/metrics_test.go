package metrics

import (
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"sync"
	"testing"
	"time"
)

func metricsInit() {
	logs.Level = logs.LevelDebug
	// To close the metrics, just remove the Init function
	Init(
		WithMetricsLog(),
		WithFlushInterval(10*time.Second),
	)
}

func StoreReport() {
	for i := 0; i < 100000; i++ {
		Store("request.store", 200, "type:test_metrics3")
		Store("request.store", 100, "type:test_metrics4")
		time.Sleep(100 * time.Millisecond)
	}
}

// test demo for store report
func TestStoreReport(t *testing.T) {
	metricsInit()
	StoreReport()
}

func CounterReport() {
	for i := 0; i < 100000; i++ {
		Counter("request.counter", 1, "type:test_metrics3")
		Counter("request.counter", 1, "type:test_metrics4")
		time.Sleep(200 * time.Millisecond)
	}
}

// test demo for counter report
func TestCounterReport(t *testing.T) {
	metricsInit()
	CounterReport()
}

func TimerReport() {
	for i := 0; i < 100000; i++ {
		begin := time.Now()
		time.Sleep(time.Duration(100) * time.Millisecond)
		Latency("request.timer", begin, "type:test_metrics3")
		begin = time.Now()
		time.Sleep(time.Duration(150) * time.Millisecond)
		Latency("request.timer", begin, "type:test_metrics4")
	}
}

// test demo for timer report
func TestTimerReport(t *testing.T) {
	metricsInit()
	TimerReport()
}

func TestReportAll(t *testing.T) {
	metricsInit()
	wg := &sync.WaitGroup{}
	wg.Add(3)
	go func() {
		StoreReport()
		wg.Done()
	}()
	go func() {
		CounterReport()
		wg.Done()
	}()
	go func() {
		TimerReport()
		wg.Done()
	}()
	wg.Wait()
}
