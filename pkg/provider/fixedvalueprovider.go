package provider

import (
	"context"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type fixedValue struct {
	data types.ProviderReport
	base
}

var _ Provider = (*fixedValue)(nil)

// NewFixedValueProvider creates fixed value provider which upon calling Provide
// will always provide the same telemetry report.
func NewFixedValueProvider(name string, data types.ProviderReport) (Provider, error) {
	return &fixedValue{
		data: data,
		base: base{
			name: name,
			kind: "fixed-value",
		},
	}, nil
}

// Provide provides the configure, fixed value report.
func (p *fixedValue) Provide(ctx context.Context) (types.ProviderReport, error) {
	return p.data, nil
}
