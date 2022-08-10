package log

import "github.com/sirupsen/logrus"

// TODO: Address logging levels and library to be used.
// See: https://github.com/Kong/kubernetes-ingress-controller/issues/1893
const (
	logrusrDiff = 4

	// InfoLevel is the converted logging level from logrus to go-logr for
	// information level logging. Note that the logrusr middleware technically
	// flattens all levels prior to this level into this level as well.
	InfoLevel = int(logrus.InfoLevel) - logrusrDiff

	// DebugLevel is the converted logging level from logrus to go-logr for
	// debug level logging.
	DebugLevel = int(logrus.DebugLevel) - logrusrDiff

	// WarnLevel is the converted logrus level to go-logr for warnings.
	WarnLevel = int(logrus.WarnLevel) - logrusrDiff
)
