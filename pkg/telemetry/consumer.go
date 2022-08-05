package telemetry

import (
	"sync"

	"github.com/go-logr/logr"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type consumer struct {
	logger logr.Logger
	once   sync.Once
	ch     chan types.Report
	done   chan struct{}
}

// Forwarder is used to forward telemetry reports to configured destination(s).
type Forwarder interface {
	Name() string
	Forward([]byte) error
}

// NewConsumer creates a new consumer which will use the provided serializer to
// serialize the data and then forward it using the provided forwarder.
func NewConsumer(s Serializer, f Forwarder) *consumer {
	var (
		ch     = make(chan types.Report)
		done   = make(chan struct{})
		logger = defaultLogger() // TODO: allow configuration
	)

	go func() {
		for {
			select {
			case <-done:
				return
			case r := <-ch:
				b, err := s.Serialize(r)
				if err != nil {
					logger.Error(err, "failed to serialize report")
					continue
				}

				if err := f.Forward(b); err != nil {
					logger.Error(err, "failed to forward report using forwarder: %s", f.Name())
				}
			}
		}
	}()

	return &consumer{
		logger: logger,
		ch:     ch,
		done:   done,
	}
}

func (c *consumer) Intake() chan<- types.Report {
	return c.ch
}

func (c *consumer) Close() {
	c.once.Do(func() {
		close(c.done)
	})
}
