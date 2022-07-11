package provider

import (
	"context"
)

type functor struct {
	f ReportFunctor
	base
}

// ReportFunctor defines a function type that functor provider accepts as means
// for delivering telemetry data.
type ReportFunctor func() (Report, error)

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

// Provide returns the Report as returned by the configured functor.
func (p *functor) Provide(ctx context.Context) (Report, error) {
	return p.f()
}
