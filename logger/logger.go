package logger

import (
	"fmt"
	"log"
	"runtime"

	// Imports the Stackdriver Logging client package.
	gce "cloud.google.com/go/logging"
	"golang.org/x/net/context"
)

// Listener Listen to log when ERROR or PANIC
type Listener func(msg string)

var (
	isLoggerGCE  bool
	loggerClient *gce.Client
	loggerInfo   *log.Logger
	loggerWarn   *log.Logger
	loggerError  *log.Logger
	loggerPanic  *gce.Logger
	logName      string

	listeners []Listener
)

func init() {
	log.Print("DJFKLJDLFJLD")
	ctx := context.Background()

	// Sets your Google Cloud Platform project ID.
	projectID := "ticklemeta-203110"

	// Sets the name of the log to write to.
	logName = "StockWatcher"

	// Checks OS
	// If Linux, use GCP Logger
	// Else, use only stdout
	if runtime.GOOS == "linux" {
		// Creates a client.
		client, err := gce.NewClient(ctx, projectID)
		if err != nil {
			loggerInfo = nil
			loggerWarn = nil
			loggerError = nil
			loggerPanic = nil
			isLoggerGCE = false
			log.Printf("Failed to create client: %v, using builtin logger", err)
			return
		}
		loggerClient = client

		loggerInfo = client.Logger(logName).StandardLogger(gce.Info)
		loggerWarn = client.Logger(logName).StandardLogger(gce.Warning)
		loggerError = client.Logger(logName).StandardLogger(gce.Error)
		loggerPanic = client.Logger(logName)

		isLoggerGCE = true
	} else {
		isLoggerGCE = false
	}

	// Logs "hello world", log entry is visible at
	// Stackdriver Logs.
	Info("Logger init")
}

// IsLoggerGCE provides interface if current logger is GCE
func IsLoggerGCE() bool {
	return isLoggerGCE
}

// Close closes GCE Client
func Close() {
	if loggerClient != nil {
		loggerClient.Close()
	}
}

// Info prints logs as this format: [INFO]
func Info(format string, v ...interface{}) {
	handleLog(loggerInfo, "INFO", format, v...)
}

// Warn prints logs as this format: [WARN]
func Warn(format string, v ...interface{}) {
	handleLog(loggerWarn, "WARN", format, v...)
}

// Error prints logs as this format: [ERROR]
func Error(format string, v ...interface{}) {
	handleLog(loggerError, "ERROR", format, v...)
}

// Errorf prints logs as this format: [ERROR]
func Errorf(format string, v error) {
	handleLog(loggerError, "ERROR", format, v.Error())
}

// Panic prints logs as this format: [PANIC]
func Panic(format string, v ...interface{}) {
	handlePanicLog(format, v...)
}

func handleLog(logHandle *log.Logger, severity, format string, v ...interface{}) {
	msgFormat := "[" + logName + "][" + severity + "] " + format

	checkExtrasStdLogPrint := func(msgFormatStr string, vv ...interface{}) {
		msg := fmt.Sprintf(msgFormatStr, vv...)
		if severity == "ERROR" {
			broadcast(msg)
		}
		log.Printf(msg)
	}

	checkExtrasGCEPrint := func(internalLogHandle *log.Logger, msgFormatStr string, vv ...interface{}) {
		msg := fmt.Sprintf(msgFormatStr, vv...)
		if severity == "ERROR" {
			broadcast(msg)
		}
		internalLogHandle.Printf(msg)
	}

	// Log to Stdout
	if logHandle == nil {
		checkExtrasStdLogPrint(msgFormat, v...)
	} else {
		checkExtrasGCEPrint(logHandle, msgFormat, v...)
	}
}

func handlePanicLog(format string, v ...interface{}) {
	const severity = gce.Critical
	msgFormat := "[" + logName + "][PANIC] " + format
	msg := fmt.Sprintf(msgFormat, v...)

	if loggerPanic == nil {
		log.Panicf(msg)
		return
	}

	s := fmt.Sprintf(format, v...)
	loggerPanic.Log(gce.Entry{
		Severity: severity,
		Payload:  s,
	})
	loggerPanic.Flush()

	panic("Killed by logger.handlePanicLog")
}

// Listen listen to ERROR and PANIC
func Listen(listener func(msg string)) {
	listeners = append(listeners, listener)
}

func broadcast(msg string) {
	for _, listener := range listeners {
		listener(msg)
	}
}
