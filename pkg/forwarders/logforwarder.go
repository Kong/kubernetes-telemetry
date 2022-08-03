package forwarders

import (
	"github.com/go-logr/logr"
)

type logForwarder struct {
	log logr.Logger
}

// NewLogForwarder creates new logForwarded which uses the provided logger to
// print all the received data.
func NewLogForwarder(log logr.Logger) *logForwarder {
	return &logForwarder{
		log: log,
	}
}

func (lf *logForwarder) Name() string {
	return "LogForwarder"
}

func (lf *logForwarder) Forward(payload []byte) error {
	lf.log.Info("got a report", "report", payload)
	return nil
}
