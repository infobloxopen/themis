package server

import (
	"github.com/Sirupsen/logrus"
)

type LevelCheckable interface {
	GetLevel() logrus.Level
}

func GetLogLevel(logger logrus.FieldLogger) logrus.Level {
	if lgr, ok := logger.(*logrus.Logger); ok {
		return lgr.Level
	} else if lgr, ok := logger.(LevelCheckable); ok {
		return lgr.GetLevel()
	}
	// else
	return 0 // return the lowest verbosity to be safe
}
