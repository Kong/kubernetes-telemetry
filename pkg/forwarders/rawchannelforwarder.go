package forwarders

import (
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type rawChannelForwarder struct {
	ch chan types.Report
}

// NewRawChannelForwarder creates new rawChannelForwarder.
func NewRawChannelForwarder(ch chan types.Report) *rawChannelForwarder {
	return &rawChannelForwarder{
		ch: ch,
	}
}

func (f *rawChannelForwarder) Name() string {
	return "LogForwarder"
}

func (f *rawChannelForwarder) Forward(r types.Report) error {
	f.ch <- r
	return nil
}
