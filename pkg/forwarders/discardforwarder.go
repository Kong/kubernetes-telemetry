package forwarders

import "context"

type discardForwarder struct{}

// NewDiscardForwarder creates a new discardForwarder which discards all received
// data.
func NewDiscardForwarder() *discardForwarder {
	return &discardForwarder{}
}

func (df *discardForwarder) Name() string {
	return "DiscardForwarder"
}

func (df *discardForwarder) Forward(ctx context.Context, payload []byte) error {
	return nil
}
