package provider

import (
	"time"
)

// NewUptimeProvider provides new uptime provider which will return uptime counted
// since the provider creation time.
func NewUptimeProvider(name string) (Provider, error) {
	start := time.Now()
	return &functor{
		f: func() (Report, error) {
			return Report{
				"uptime": int(time.Since(start).Truncate(time.Second).Seconds()),
			}, nil
		},
		base: base{
			name: name,
			kind: "uptime",
		},
	}, nil
}
