package metrics

import (
	"github.com/byteplus-sdk/sdk-go/core/logs"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func metricsInit() {
	SetPrintLog(true)
	logs.Level = logs.LevelDebug
}

func StoreReport()  {
	report := NewReporterBuilder().
		EnableMetrics(true).BaseTags(map[string]string{"tenant": "metrics_demo"}).Build()
	for i := 0; i < 100000; i++ {
		report.Store("request.store", 200, "type:test_metrics1")
		report.Store("request.store", 100, "type:test_metrics2")
		time.Sleep(100 * time.Millisecond)
	}
}

// test demo for store report
func TestStoreReport(t *testing.T) {
	metricsInit()
	StoreReport()
}

func CounterReport()  {
	report := NewReporterBuilder().
		EnableMetrics(true).BaseTags(map[string]string{"tenant": "metrics_demo"}).Build()
	for i := 0; i < 100000; i++ {
		report.Counter("request.qps", 1, "type:test_metrics1")
		report.Counter("request.qps", 1, "type:test_metrics2")
		time.Sleep(100 * time.Millisecond)
	}
}

// test demo for counter report
func TestCounterReport(t *testing.T) {
	metricsInit()
	CounterReport()
}

func TimerReport()  {
	report := NewReporterBuilder().
		EnableMetrics(true).BaseTags(map[string]string{"tenant": "metrics_demo"}).Build()
	for i := 0; i < 100000; i++ {
		begin := time.Now()
		time.Sleep(time.Duration(rand.Int31n(100)) * time.Millisecond)
		report.Latency("request.latency", begin, "type:test_metrics1")
		begin = time.Now()
		time.Sleep(time.Duration(rand.Int31n(150)) * time.Millisecond)
		report.Latency("request.latency", begin, "type:test_metrics2")
	}
}

// test demo for timer report
func TestTimerReport(t *testing.T) {
	metricsInit()
	TimerReport()
}

func TestReportAll(t *testing.T)  {
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
