package provider

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {
	testcases := []struct {
		name     string
		r1, r2   Report
		expected Report
	}{
		{
			name: "basic",
			r1: Report{
				ClusterArchKey: "arm64",
			},
			r2: Report{
				ClusterVersionKey: "1.23.0",
			},
			expected: Report{
				ClusterArchKey:    "arm64",
				ClusterVersionKey: "1.23.0",
			},
		},
		{
			name: "merged in value overwrites what's already in the report",
			r1: Report{
				ClusterArchKey: "arm64",
			},
			r2: Report{
				ClusterArchKey:    "amd64",
				ClusterVersionKey: "1.23.0",
			},
			expected: Report{
				ClusterArchKey:    "amd64",
				ClusterVersionKey: "1.23.0",
			},
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.r1.Merge(tc.r2)
			require.EqualValues(t, tc.expected, tc.r1)
		})
	}
}
