package metrics

import (
	"fmt"
	"testing"
	"time"

	"github.com/byteplus-sdk/byteplus-sdk-go-rec-core/logs"
	"github.com/google/uuid"
)

var times = 100

func metricsInit() {
	logs.Level = logs.LevelDebug
	Collector.InitWithOptions(
		EnableMetrics(),
		EnableMetricsLog(),
		WithMetricsHTTPSchema("http"),
		WithMetricsDomain("rec-b-ap-singapore-1.byteplusapi.com"),
		WithMetricsPrefix("test.byteplus.sdk"),
		WithMetricsTimeout(time.Second*3),
		WithReportInterval(time.Second*5),
	)
}

// test demo for store report
func TestStoreReport(t *testing.T) {
	metricsInit()
	StoreReport()
}

func StoreReport() {
	fmt.Println("start store reporting...")
	for i := 0; i < times; i++ {
		Store("request.store", 200, "type:test_metrics3")
		Store("request.store", 100, "type:test_metrics4")
		time.Sleep(5 * time.Second)
	}
	fmt.Println("stop store reporting")
}

// test demo for counter report
func TestCounterReport(t *testing.T) {
	metricsInit()
	CounterReport()
}

func CounterReport() {
	fmt.Println("start counter reporting...")
	for i := 0; i < times; i++ {
		Counter("request.counter", 1, "type:test_metrics1")
		Counter("request.counter", 1, "type:test_metrics2")
		time.Sleep(20 * time.Second)
	}
	fmt.Println("stop counter reporting")
}

// test demo for timerValue report
func TestTimerReport(t *testing.T) {
	metricsInit()
	TimerReport()
}

func TimerReport() {
	fmt.Println("start timer reporting...")
	for i := 0; i < times; i++ {
		Timer("request.timer", 140, "type:test_metrics3")
		Timer("request.timer", 160, "type:test_metrics3")
		time.Sleep(30 * time.Second)
	}
	fmt.Println("stop timer reporting")
}

// test demo for metrics log report
func TestMetricsLogReport(t *testing.T) {
	metricsInit()
	logID := uuid.NewString()
	fmt.Printf("logID: %s\n", logID)
	Error(uuid.NewString(), "this is a test log, name:%s, num: %d", "demo", 2)
	time.Sleep(time.Second * 20)
}
