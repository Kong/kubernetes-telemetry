package provider

import (
	"context"
	"time"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// UptimeKey is the report key that under which one can find uptime.
	UptimeKey = types.ProviderReportKey("uptime")
)

// NewUptimeProvider provides new uptime provider which will return uptime counted
// since the provider creation time.
func NewUptimeProvider(name string) (Provider, error) {
	start := time.Now()
	return &functor{
		f: func(ctx context.Context) (types.ProviderReport, error) {
			return types.ProviderReport{
				UptimeKey: int(time.Since(start).Truncate(time.Second).Seconds()),
			}, nil
		},
		base: base{
			name: name,
			kind: "uptime",
		},
	}, nil
}
