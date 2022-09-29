package metrics

import "fmt"

func Trace(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelTrace, currentTimeMillis())
}

func Debug(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelDebug, currentTimeMillis())
}

func Info(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelInfo, currentTimeMillis())
}

func Notice(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelNotice, currentTimeMillis())
}

func Warn(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelWarn, currentTimeMillis())
}

func Error(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelError, currentTimeMillis())
}

func Fatal(logID, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Collector.EmitLog(logID, message, logLevelFatal, currentTimeMillis())
}
