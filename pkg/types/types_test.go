package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReport(t *testing.T) {
	testcases := []struct {
		name     string
		r1, r2   ProviderReport
		expected ProviderReport
	}{
		{
			name: "basic",
			r1: ProviderReport{
				ProviderReportKey("ArchKey"): "arm64",
			},
			r2: ProviderReport{
				ProviderReportKey("VersionKey"): "1.23.0",
			},
			expected: ProviderReport{
				ProviderReportKey("ArchKey"):    "arm64",
				ProviderReportKey("VersionKey"): "1.23.0",
			},
		},
		{
			name: "merged in value overwrites what's already in the report",
			r1: ProviderReport{
				ProviderReportKey("ArchKey"): "arm64",
			},
			r2: ProviderReport{
				ProviderReportKey("ArchKey"):    "amd64",
				ProviderReportKey("VersionKey"): "1.23.0",
			},
			expected: ProviderReport{
				ProviderReportKey("ArchKey"):    "amd64",
				ProviderReportKey("VersionKey"): "1.23.0",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.r1.Merge(tc.r2)
			require.EqualValues(t, tc.expected, tc.r1)
		})
	}
}
