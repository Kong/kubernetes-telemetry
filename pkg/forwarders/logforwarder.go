package forwarders

import (
	"context"

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

// Name returns the name of the forwarder.
func (lf *logForwarder) Name() string {
	return "LogForwarder"
}

// Forward logs the received payload.
func (lf *logForwarder) Forward(ctx context.Context, payload []byte) error {
	lf.log.Info("got a report", "report", payload)
	return nil
}
