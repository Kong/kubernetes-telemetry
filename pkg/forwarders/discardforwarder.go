package forwarders

import "context"

type discardForwarder struct{}

// NewDiscardForwarder creates a new discardForwarder which discards all received
// data.
func NewDiscardForwarder() *discardForwarder {
	return &discardForwarder{}
}

// Name returns the name of the forwarder.
func (df *discardForwarder) Name() string {
	return "DiscardForwarder"
}

// Forward discards the received payload.
func (df *discardForwarder) Forward(ctx context.Context, payload []byte) error {
	return nil
}
