package forwarders

import (
	"context"
)

type channelForwarder struct {
	ch chan []byte
}

// NewChannelForwarder creates new channelForwarder.
func NewChannelForwarder(ch chan []byte) *channelForwarder {
	return &channelForwarder{
		ch: ch,
	}
}

// Name returns the name of the forwarder.
func (f *channelForwarder) Name() string {
	return "channelForwarder"
}

// Forward forwards the received report on the configured channel.
func (f *channelForwarder) Forward(ctx context.Context, r []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case f.ch <- r:
	}

	return nil
}
