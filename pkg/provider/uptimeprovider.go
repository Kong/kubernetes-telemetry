package provider

import (
	"time"
)

const (
	// UptimeKey is the report key that under which one can find uptime.
	UptimeKey = ReportKey("uptime")
)

// NewUptimeProvider provides new uptime provider which will return uptime counted
// since the provider creation time.
func NewUptimeProvider(name string) (Provider, error) {
	start := time.Now()
	return &functor{
		f: func() (Report, error) {
			return Report{
				UptimeKey: int(time.Since(start).Truncate(time.Second).Seconds()),
			}, nil
		},
		base: base{
			name: name,
			kind: "uptime",
		},
	}, nil
}
