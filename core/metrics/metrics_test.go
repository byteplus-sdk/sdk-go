package metrics

import (
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"sync"
	"testing"
	"time"
)

var times = 1500

func metricsInit() {
	logs.Level = logs.LevelDebug
	// To close the metrics, just remove the Init function
	Init(
		WithMetricsLog(),
		WithFlushInterval(10*time.Second),
	)
}

func StoreReport(times int) {
	for i := 0; i < times; i++ {
		Store("request.store", 200, "type:test_metrics3")
		Store("request.store", 100, "type:test_metrics4")
		time.Sleep(100 * time.Millisecond)
	}
	println("stop store reporting")
}

// test demo for store report
func TestStoreReport(t *testing.T) {
	metricsInit()
	StoreReport(100000)
}

func CounterReport(times int) {
	for i := 0; i < times; i++ {
		Counter("request.counter", 1, "type:test_metrics3")
		Counter("request.counter", 1, "type:test_metrics4")
		time.Sleep(200 * time.Millisecond)
	}
	println("stop counter reporting")
}

// test demo for counter report
func TestCounterReport(t *testing.T) {
	metricsInit()
	CounterReport(1000000)
}

func TimerReport(times int) {
	for i := 0; i < times; i++ {
		begin := time.Now()
		time.Sleep(time.Duration(100) * time.Millisecond)
		Latency("request.timer", begin, "type:test_metrics3")
		begin = time.Now()
		time.Sleep(time.Duration(150) * time.Millisecond)
		Latency("request.timer", begin, "type:test_metrics4")
	}
	println("stop timer reporting")
}

// test demo for timerValue report
func TestTimerReport(t *testing.T) {
	metricsInit()
	TimerReport(1000000)
}

func TestReportAll(t *testing.T) {
	metricsInit()
	wg := &sync.WaitGroup{}
	goNum := 10
	for i := 0; i < goNum; i++ {
		wg.Add(3)
		go func() {
			StoreReport(times)
			time.Sleep(100 * time.Second)
			wg.Done()
		}()
		go func() {
			CounterReport(times)
			time.Sleep(100 * time.Second)
			wg.Done()
		}()
		go func() {
			TimerReport(times)
			time.Sleep(100 * time.Second)
			wg.Done()
		}()
	}

	wg.Wait()
}
