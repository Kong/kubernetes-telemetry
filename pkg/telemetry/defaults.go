package telemetry

import (
	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func defaultLogger() logr.Logger {
	return logrusr.New(logrus.New())
}
