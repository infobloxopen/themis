package server

import (
	"github.com/Sirupsen/logrus"
	"github.com/infobloxopen/themis/pdp"
)

const lowestVerbosityLevel = logrus.PanicLevel

// LevelCheckable interface ensures public getter for log level
type LevelCheckable interface {
	GetLevel() logrus.Level
}

// DetailedInfo interface provides detailed information consumed during WithFields.
// 	- GetDetail should be an expensive operation that provides more information
// 	- FilterLevel should be checked against the logger level to ensure that details
//	  are only processed when necessary
//	- String ensures DetailedInfo provides a brief description in the event
//	  that the logger does not consume through GetDetail
type DetailedInfo interface {
	GetDetail(nShow uint) string
	FilterLevel() logrus.Level
	String() string
}

// compile-time interface check
var (
	_ DetailedInfo = pdp.PolicyUpdateDetail{}
)

func getLogLevel(logger logrus.FieldLogger) logrus.Level {
	switch log := logger.(type) {
	case *logrus.Logger:
		return log.Level
	case LevelCheckable:
		return log.GetLevel()
	}
	return lowestVerbosityLevel // return the lowest verbosity to be safe
}
