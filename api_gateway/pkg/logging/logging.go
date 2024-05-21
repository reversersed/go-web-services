package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
)

type Hook struct {
	Writer []io.Writer
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

type Logger struct {
	*logrus.Entry
}

func NewLogger() *Logger {
	return &Logger{e}
}

func (l *Logger) NewLoggerWithFields(field string, value interface{}) *Logger {
	return &Logger{l.WithField(field, value)}
}

func init() {
	l := logrus.New()
	l.SetReportCaller(true)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (string, string){
			filename := path.Base(f.File)
			return fmt.Sprintf("%s:%d", filename, f.Line), fmt.Sprintf("%s()",f.Function)
		},
		DisableColors: false,
		FullTimestamp: true,
	}

	err := os.MkdirAll("logs", 0644)
	if err != nil || os.IsExist(err){
		panic("can't create logs directory")
	}

	logFile, err := os.OpenFile("logs/all.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(fmt.Sprintf("[Message]: %s", err))
	}
	l.SetOutput(io.Discard)

	l.AddHook(&Hook {
			Writer: []io.Writer{logFile, os.Stdout}, 
			LogLevels: logrus.AllLevels,
		})

	l.SetLevel(logrus.TraceLevel)

	e = logrus.NewEntry(l)
}