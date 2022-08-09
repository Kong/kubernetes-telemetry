package provider

import (
	"os"
)

const (
	// HostnameKey is the report key that under which one can find hostname.
	HostnameKey = ReportKey("hn")
)

// NewHostnameProvider creates hostname provider.
func NewHostnameProvider(name string) (Provider, error) {
	return &functor{
		f: func() (Report, error) {
			hostname, err := os.Hostname()
			if err != nil {
				return nil, err
			}
			return Report{
				HostnameKey: hostname,
			}, nil
		},
		base: base{
			name: name,
			kind: "hostname",
		},
	}, nil
}
