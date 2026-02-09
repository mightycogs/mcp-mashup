package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mightycogs/mcp-mashup/pkg/config"
)

var (
	logFile  *os.File
	errorLog *log.Logger
	infoLog  *log.Logger
	debugLog *log.Logger
	traceLog *log.Logger
	logLevel config.LogLevel
	initOnce sync.Once
)

// Init initializes the logger with the specified log level and optional log file
func Init(level config.LogLevel, logFilePath string) error {
	var err error
	initOnce.Do(func() {
		logLevel = level

		var logWriter io.Writer
		if logFilePath != "" {
			if err = os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
				err = fmt.Errorf("failed to create log directory: %w", err)
				return
			}

			logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				err = fmt.Errorf("failed to open log file: %w", err)
				return
			}
			logWriter = logFile
		} else {
			logWriter = os.Stderr
		}

		errorLog = log.New(logWriter, "ERROR: ", log.Ldate|log.Ltime)
		infoLog = log.New(logWriter, "INFO: ", log.Ldate|log.Ltime)
		debugLog = log.New(logWriter, "DEBUG: ", log.Ldate|log.Ltime)
		traceLog = log.New(logWriter, "TRACE: ", log.Ldate|log.Ltime)

		Info("Logger initialized with level: %v, log file: %v", level, logFilePath)
	})
	return err
}

// Close closes the log file if one is open
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func Error(format string, v ...interface{}) {
	errorLog.Printf(format, v...)
}

func Info(format string, v ...interface{}) {
	if logLevel >= config.LogLevelInfo {
		infoLog.Printf(format, v...)
	}
}

func Debug(format string, v ...interface{}) {
	if logLevel >= config.LogLevelDebug {
		debugLog.Printf(format, v...)
	}
}

func Trace(format string, v ...interface{}) {
	if logLevel >= config.LogLevelTrace {
		traceLog.Printf(format, v...)
	}
}

// LogRequest logs incoming JSON-RPC requests
func LogRequest(method string, id interface{}, params interface{}) {
	if logLevel >= config.LogLevelDebug {
		debugLog.Printf("Request: method=%s, id=%v", method, id)
		if logLevel >= config.LogLevelTrace {
			traceLog.Printf("Request params: %+v", params)
		}
	}
}

// LogResponse logs outgoing JSON-RPC responses
func LogResponse(id interface{}, result interface{}, err error) {
	if logLevel >= config.LogLevelDebug {
		if err != nil {
			debugLog.Printf("Response: id=%v, error=%v", id, err)
		} else {
			debugLog.Printf("Response: id=%v, success=true", id)
			if logLevel >= config.LogLevelTrace {
				traceLog.Printf("Response result: %+v", result)
			}
		}
	}
}

func LogRPC(direction string, message []byte) {
	if logLevel >= config.LogLevelTrace {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		traceLog.Printf("%s RPC [%s]: %s", direction, timestamp, string(message))

		var jsonMsg map[string]interface{}
		if err := json.Unmarshal(message, &jsonMsg); err == nil {
			prettyJSON, err := json.MarshalIndent(jsonMsg, "", "  ")
			if err == nil {
				traceLog.Printf("%s RPC PARSED [%s]:\n%s", direction, timestamp, string(prettyJSON))
			}
		}
	}
}

func Fatal(format string, v ...interface{}) {
	if errorLog != nil {
		errorLog.Printf(format, v...)
	}
	fmt.Fprintf(os.Stderr, "FATAL: "+format+"\n", v...)
	Close()
	os.Exit(1)
}
