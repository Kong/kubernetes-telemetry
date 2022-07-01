package pkg

import (
	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

func DefaultLogger() logr.Logger {
	return logrusr.New(logrus.New())
}
