package forwarders

import (
	"context"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type rawChannelForwarder struct {
	ch chan types.SignalReport
}

// NewRawChannelForwarder creates new rawChannelForwarder.
func NewRawChannelForwarder(ch chan types.SignalReport) *rawChannelForwarder {
	return &rawChannelForwarder{
		ch: ch,
	}
}

// Name returns the name of the forwarder.
func (f *rawChannelForwarder) Name() string {
	return "rawChannelForwarder"
}

// Forward forwards the received report on the configured channel.
func (f *rawChannelForwarder) Forward(ctx context.Context, r types.SignalReport) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case f.ch <- r:
	}

	return nil
}
