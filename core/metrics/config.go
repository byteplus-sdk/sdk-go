package metrics

// Some global config of metrics should be set here

var (
	metricsDomain string = defaultMetricsDomain
	printLog      bool //whether print logs when during collecting metrics
)

func SetMetricsDomain(domain string)  {
	if metricsDomain != "" {
		metricsDomain = domain
	}
}

func SetPrintLog(enableLog bool)  {
	printLog = enableLog
}

func GetMetricsDomain() string {
	return metricsDomain
}

func IsEnablePrintLog() bool {
	return printLog
}
