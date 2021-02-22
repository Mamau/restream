package storage

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
	"sync"
)

type LogType string

const (
	INFO    LogType = "INFO"
	WARNING LogType = "WARNING"
	FATAL   LogType = "FATAL"
	ERROR   LogType = "ERROR"
)

type StreamLogger struct {
	warningLogger *log.Logger
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	fatalLogger   *log.Logger
}

var once sync.Once
var instance *StreamLogger

func GetLogger() *StreamLogger {
	once.Do(func() {
		instance = &StreamLogger{
			infoLogger:    log.New(os.Stdout, fmt.Sprintf("[%s]: ", INFO), log.Ldate|log.Ltime),
			warningLogger: log.New(os.Stdout, fmt.Sprintf("[%s]: ", WARNING), log.Ldate|log.Ltime),
			errorLogger:   log.New(os.Stderr, fmt.Sprintf("[%s]: ", ERROR), log.Ldate|log.Ltime),
			fatalLogger:   log.New(os.Stderr, fmt.Sprintf("[%s]: ", FATAL), log.Ldate|log.Ltime),
		}
	})
	return instance
}

func (l *StreamLogger) Info(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.infoLogger.Println(message)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelInfo)
	})
	sentry.CaptureMessage(message)
}

func (l *StreamLogger) Warning(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.warningLogger.Println(message)

	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelWarning)
	})
	sentry.CaptureMessage(message)
}

func (l *StreamLogger) Error(e error) {
	l.errorLogger.Println(e.Error())
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
	})
	sentry.CaptureException(e)
}

func (l *StreamLogger) Fatal(e error) {
	l.fatalLogger.Println(e.Error())
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelFatal)
	})
	sentry.CaptureException(e)
}
