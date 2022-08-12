package telemetry

import (
	"context"
	"sync"

	"github.com/go-logr/logr"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type consumer struct {
	logger logr.Logger
	once   sync.Once
	ch     chan types.Report
	cancel func()
}

// Forwarder is used to forward telemetry reports to configured destination(s).
type Forwarder interface {
	Name() string
	Forward(context.Context, []byte) error
}

// NewConsumer creates a new consumer which will use the provided serializer to
// serialize the data and then forward it using the provided forwarder.
func NewConsumer(s Serializer, f Forwarder) *consumer {
	var (
		ch          = make(chan types.Report)
		ctx, cancel = context.WithCancel(context.Background())
		// TODO: allow configuration: https://github.com/Kong/kubernetes-telemetry/issues/46
		logger = defaultLogger()
	)

	go func() {
		done := ctx.Done()

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

				if err := f.Forward(ctx, b); err != nil {
					logger.Error(err, "failed to forward report", "forwarder", f.Name())
				}
			}
		}
	}()

	return &consumer{
		logger: logger,
		ch:     ch,
		cancel: cancel,
	}
}

// Intake returns a channel on which this consumer will wait for data to consume it.
func (c *consumer) Intake() chan<- types.Report {
	return c.ch
}

// Close closes the consumer.
func (c *consumer) Close() {
	c.once.Do(func() {
		c.cancel()
	})
}

type rawConsumer struct {
	logger logr.Logger
	once   sync.Once
	ch     chan types.Report
	cancel func()
}

// RawForwarder is used to forward raw, unserialized telemetry reports to configured
// destination(s).
type RawForwarder interface {
	Name() string
	Forward(context.Context, types.Report) error
}

// NewRawConsumer creates a new rawconsumer that will use the provided raw forwarder
// to forward received reports.
func NewRawConsumer(f RawForwarder) *rawConsumer {
	var (
		ch          = make(chan types.Report)
		ctx, cancel = context.WithCancel(context.Background())
		// TODO: allow configuration: https://github.com/Kong/kubernetes-telemetry/issues/46
		logger = defaultLogger()
	)

	go func() {
		done := ctx.Done()

		for {
			select {
			case <-done:
				return
			case r := <-ch:
				if err := f.Forward(ctx, r); err != nil {
					logger.Error(err, "failed to forward report using raw forwarder: %s", f.Name())
				}
			}
		}
	}()

	return &rawConsumer{
		logger: logger,
		ch:     ch,
		cancel: cancel,
	}
}

// Intake returns a channel on which this consumer will wait for data to consume it.
func (c *rawConsumer) Intake() chan<- types.Report {
	return c.ch
}

// Close closes rawconsumer.
func (c *rawConsumer) Close() {
	c.once.Do(func() {
		c.cancel()
	})
}
