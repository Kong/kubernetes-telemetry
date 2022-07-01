package provider

import (
	"context"
)

// Provider defines how a telemetry provider can be used.
type Provider interface {
	Name() string
	Kind() string
	Provide(context.Context) (Report, error)
}
