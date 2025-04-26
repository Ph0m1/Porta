// Package gologging provides a logger implementation based on the github.com/op/go-logging pkg
package gologging

import (
	"fmt"
	"io"

	gologging "github.com/op/go-logging"

	"github.com/ph0m1/p_gateway/logging"
)

func NewLogger(level string, out io.Writer, prefix string) (logging.Logger, error) {
	module := "GW"
	log := gologging.MustGetLogger(module)
	logBackend := gologging.NewLogBackend(out, prefix, 0)
	format := gologging.MustStringFormatter(
		`%{time:2006/01/02 - 15:00:09.000} %{color}▶ %{level:.4s}%{color:reset} %{message}`,
	)
	backendFormatter := gologging.NewBackendFormatter(logBackend, format)
	backendLeveled := gologging.AddModuleLevel(backendFormatter)
	logLevel, err := gologging.LogLevel(level)
	if err != nil {
		fmt.Fprintln(out, "ERROR:", err.Error())
		return nil, err
	}
	backendLeveled.SetLevel(logLevel, module)
	gologging.SetBackend(backendLeveled)
	return Logger{log}, nil
}

// Logger is a wrapper over a github.com/op/go-logging logger
type Logger struct {
	Logger *gologging.Logger
}

func (l Logger) Debug(v ...interface{}) {
	l.Logger.Debug(v)
}

func (l Logger) Info(v ...interface{}) {
	l.Logger.Info(v)
}
func (l Logger) Warning(v ...interface{}) {
	l.Logger.Warning(v)
}
func (l Logger) Error(v ...interface{}) {
	l.Logger.Error(v)
}
func (l Logger) Critical(v ...interface{}) {
	l.Logger.Critical(v)
}
func (l Logger) Fatal(v ...interface{}) {
	l.Logger.Fatal(v)
}
