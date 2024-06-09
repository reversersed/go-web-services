package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	dateTimeLayout = "2006-01-02 15.04.05"
)

type Hook struct {
	Writer    []io.Writer
	LogLevels []logrus.Level
}

func (hook *Hook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	for _, w := range hook.Writer {
		_, err = w.Write([]byte(line))
	}
	if err != nil {
		return err
	}
	return nil
}
func (hook *Hook) Levels() []logrus.Level {
	return hook.LogLevels
}

var e *logrus.Entry
var once sync.Once

type Logger struct {
	*logrus.Entry
}

func GetLogger() *Logger {
	once.Do(func() {
		l := logrus.New()
		l.SetReportCaller(true)
		l.Formatter = &logrus.TextFormatter{
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()", f.Function)
			},
			DisableColors: false,
			FullTimestamp: true,
		}

		err := os.MkdirAll("logs", 0644)
		if err != nil || os.IsExist(err) {
			panic("can't create logs directory")
		}
		err = os.MkdirAll("logs/debug", 0644)
		if err != nil || os.IsExist(err) {
			panic("can't create logs directory")
		}
		err = os.MkdirAll("logs/errors", 0644)
		if err != nil || os.IsExist(err) {
			panic("can't create logs directory")
		}
		err = os.MkdirAll("logs/warnings", 0644)
		if err != nil || os.IsExist(err) {
			panic("can't create logs directory")
		}
		t := time.Now().UTC().Format(dateTimeLayout)
		logFile, err := os.OpenFile(fmt.Sprintf("logs/debug/%s.log", t), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			panic(fmt.Sprintf("[Message]: %s", err))
		}

		errorLogFile, err := os.OpenFile(fmt.Sprintf("logs/errors/%s.log", t), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			panic(fmt.Sprintf("[Message]: %s", err))
		}
		warnLogFile, err := os.OpenFile(fmt.Sprintf("logs/warnings/%s.log", t), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
		if err != nil {
			panic(fmt.Sprintf("[Message]: %s", err))
		}
		l.SetOutput(io.Discard)

		l.AddHook(&Hook{
			Writer:    []io.Writer{logFile, os.Stdout},
			LogLevels: []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.TraceLevel},
		})
		l.AddHook(&Hook{
			Writer:    []io.Writer{errorLogFile, os.Stdout},
			LogLevels: []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel},
		})
		l.AddHook(&Hook{
			Writer:    []io.Writer{warnLogFile, os.Stdout},
			LogLevels: []logrus.Level{logrus.WarnLevel},
		})

		l.SetLevel(logrus.TraceLevel)

		e = logrus.NewEntry(l)
	})
	return &Logger{e}
}

func (l *Logger) NewLoggerWithFields(field string, value interface{}) *Logger {
	return &Logger{l.WithField(field, value)}
}
