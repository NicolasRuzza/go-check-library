package logger

type LogType string

const (
	INFO    LogType = "INFO"
	WARN    LogType = "WARN"
	ERROR   LogType = "ERROR"
	SUCCESS LogType = "SUCCESS"
)
