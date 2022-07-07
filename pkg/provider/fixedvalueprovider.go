package provider

import (
	"context"
)

type fixedValue struct {
	data Report
	base
}

var _ Provider = (*fixedValue)(nil)

// NewFixedValueProvider creates fixed value provider which upon calling Provide
// will always provide the same telemetry report.
func NewFixedValueProvider(name string, data Report) (Provider, error) {
	return &fixedValue{
		data: data,
		base: base{
			name: name,
			kind: "fixed-value",
		},
	}, nil
}

// Provide provides the configure, fixed value report.
func (p *fixedValue) Provide(ctx context.Context) (Report, error) {
	return p.data, nil
}
