package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debugf(fmt string, args ...interface{})
	Printf(fmt string, args ...interface{})
	Warningf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Fatalf(fmt string, args ...interface{})
	SetOutput(w io.Writer)
}

type Log struct {
	name string
	l    *logrus.Logger
}

func New(name, logLevel string) Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	if level, err := logrus.ParseLevel(logLevel); err == nil {
		l.SetLevel(level)
	}
	return &Log{name: name, l: l}
}

func (l *Log) SetOutput(w io.Writer) {
	l.l.SetOutput(w)
}

func (l *Log) Debugf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Debugf(fmt, args...)
}

func (l *Log) Printf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Printf(fmt, args...)
}

func (l *Log) Warningf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Warningf(fmt, args...)
}

func (l *Log) Errorf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Errorf(fmt, args...)
}

func (l *Log) Fatalf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Fatalf(fmt, args...)
}
