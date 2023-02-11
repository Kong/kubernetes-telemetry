package provider

import (
	"context"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type functor struct {
	f ReportFunctor
	base
}

// ReportFunctor defines a function type that functor provider accepts as means
// for delivering telemetry data.
type ReportFunctor func(context.Context) (types.ProviderReport, error)

var _ Provider = (*functor)(nil)

// NewFunctorProvider creates a new functor provider that allows to define one's
// own telemetry retrieval logic by providing a ReportFunctor as parameter.
func NewFunctorProvider(name string, f ReportFunctor) (Provider, error) {
	return &functor{
		f: f,
		base: base{
			name: name,
			kind: "functor",
		},
	}, nil
}

// Provide returns the types.ProviderReport as returned by the configured functor.
func (p *functor) Provide(ctx context.Context) (types.ProviderReport, error) {
	return p.f(ctx)
}
