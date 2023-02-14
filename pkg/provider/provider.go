package provider

import (
	"context"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

// Kind presents provider's kind.
type Kind string

// Provider defines how a telemetry provider can be used.
type Provider interface {
	Name() string
	Kind() Kind
	Provide(context.Context) (types.ProviderReport, error)
}
