package provider

import (
	"context"
	"os"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// HostnameKey is the report key that under which one can find hostname.
	HostnameKey = types.ProviderReportKey("hn")
)

// NewHostnameProvider creates hostname provider.
func NewHostnameProvider(name string) (Provider, error) {
	return &functor{
		f: func(ctx context.Context) (types.ProviderReport, error) {
			hostname, err := os.Hostname()
			if err != nil {
				return nil, err
			}
			return types.ProviderReport{
				HostnameKey: hostname,
			}, nil
		},
		base: base{
			name: name,
			kind: "hostname",
		},
	}, nil
}
