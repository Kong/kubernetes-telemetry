package provider

import (
	"context"
)

type functor struct {
	f ReportFunctor
	base
}

type ReportFunctor func() (Report, error)

var _ Provider = (*functor)(nil)

// NewFunctorProvider creates a new functor provider that allows to define one's
// own telemetry retrieval logic by providing a ReportFunctor as parameter.
func NewFunctorProvider(name string, f ReportFunctor) (Provider, error) {
	return functor{
		f: f,
		base: base{
			name: name,
			kind: "functor",
		},
	}, nil
}

func (p functor) Provide(ctx context.Context) (Report, error) {
	return p.f()
}
