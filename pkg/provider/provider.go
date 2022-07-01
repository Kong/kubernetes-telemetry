package provider

import (
	"context"
)

// Kind presents provider's kind.
type Kind string

// Provider defines how a telemetry provider can be used.
type Provider interface {
	Name() string
	Kind() Kind
	Provide(context.Context) (Report, error)
}
