package server

import (
	"github.com/Sirupsen/logrus"
)

const lowestVerbosityLevel = logrus.PanicLevel

// LevelCheckable interface ensures public getter for log level
type LevelCheckable interface {
	GetLevel() logrus.Level
}

func getLogLevel(logger logrus.FieldLogger) logrus.Level {
	switch log := logger.(type) {
	case *logrus.Logger:
		return log.Level
	case LevelCheckable:
		return log.GetLevel()
	}
	return lowestVerbosityLevel // return the lowest verbosity to be safe
}
