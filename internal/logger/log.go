package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Printf(fmt string, args ...interface{})
	Warningf(fmt string, args ...interface{})
	Fatalf(fmt string, args ...interface{})
}

type Log struct {
	name string
	l    *logrus.Logger
}

func New(name string) Logger {
	return &Log{
		name: name,
		l:    logrus.New(),
	}
}

func (l *Log) Printf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Printf(fmt, args...)
}

func (l *Log) Warningf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Warningf(fmt, args...)
}

func (l *Log) Fatalf(fmt string, args ...interface{}) {
	l.l.WithFields(logrus.Fields{"name": l.name}).Fatalf(fmt, args...)
}
