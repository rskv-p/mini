package log_type

// ILogClient defines methods for logging various levels of logs with additional context.
type ILogClient interface {
	Trace(message string, context map[string]interface{})
	Debug(message string, context map[string]interface{})
	Info(message string, context map[string]interface{})
	Warn(message string, context map[string]interface{})
	Error(message string, context map[string]interface{})
}
