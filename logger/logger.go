package logger

import (
	"fmt"
	"os"
)

const (
	VERB  LogLevel = "VERB"
	INFO           = "INFO"
	WARN           = "WARN"
	ERROR          = "ERROR"
)

type (
	LogLevel string

	Logger interface {
		Log(level LogLevel, format string, args ...interface{})
	}

	stdLogger struct {
		level LogLevel
	}
)

func IsLogLevelAtLeast(minLevel, checkLevel LogLevel) bool {
	logLevels := map[LogLevel]int{
		VERB:  1,
		INFO:  2,
		WARN:  3,
		ERROR: 4,
	}

	minLevelValue, ok := logLevels[minLevel]
	if !ok {
		return false
	}

	checkLevelValue, ok := logLevels[checkLevel]
	if !ok {
		return false
	}

	return minLevelValue <= checkLevelValue
}

func (s stdLogger) Log(level LogLevel, format string, args ...interface{}) {
	if !IsLogLevelAtLeast(s.level, level) {
		return
	}

	message := fmt.Sprintf(format, args...)
	_, _ = os.Stdout.WriteString(string(level) + ": " + message + "\n")
}

func ProvideLogger() Logger {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = INFO
	}
	return &stdLogger{
		level: LogLevel(logLevel),
	}
}
