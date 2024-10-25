package meshdetect

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestMeshDeploymentResultsToProviderReport(t *testing.T) {
	testCases := []struct {
		caseName string
		results  meshDeploymentResults
		expected types.ProviderReport
	}{
		{
			caseName: "deployment:kong-mesh",
			results: meshDeploymentResults{
				MeshKindKongMesh: {
					ServiceExists: true,
				},
			},
			expected: types.ProviderReport{
				"mdep": "km3",
			},
		},
		{
			caseName: "deployment:traefik",
			results: meshDeploymentResults{
				MeshKindTraefik: {
					ServiceExists: true,
				},
			},
			expected: types.ProviderReport{
				"mdep": "t3",
			},
		},
		{
			caseName: "deployment:consul,aws-app-mesh",
			results: meshDeploymentResults{
				MeshKindConsul: {
					ServiceExists: true,
				},
				MeshKindAWSAppMesh: {
					ServiceExists: true,
				},
			},
			expected: types.ProviderReport{
				"mdep": "a3,c3",
			},
		},
		{
			caseName: "deployment:nil results should produce empty report",
			results:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.results.ToProviderReport()
			require.Equal(t, tc.expected, report)
		})
	}
}

func TestMeshRunUnderResultToProviderReport(t *testing.T) {
	testCases := []struct {
		caseName string
		results  meshRunUnderResults
		expected types.ProviderReport
	}{
		{
			caseName: "run_under:istio,linkerd",
			results: meshRunUnderResults{
				MeshKindIstio: {
					PodOrServiceAnnotation: true,
				},
				MeshKindLinkerd: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    true,
				},
			},
			expected: types.ProviderReport{
				"kinm": "i2,l2,l3,l4",
			},
		},
		{
			caseName: "run_under:kuma,kong-mesh",
			results: meshRunUnderResults{
				MeshKindKuma: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
				},
				MeshKindKongMesh: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
				},
			},
			expected: types.ProviderReport{
				"kinm": "k2,k3,km2,km3",
			},
		},
		{
			caseName: "run_under:traefik,aws-app-mesh",
			results: meshRunUnderResults{
				MeshKindTraefik: {
					PodOrServiceAnnotation: true,
				},
				MeshKindAWSAppMesh: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    true,
				},
			},
			expected: types.ProviderReport{
				"kinm": "a2,a3,a4,t2",
			},
		},
		{
			caseName: "run_under:should return empty report for nil results",
			results:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.results.ToProviderReport()
			require.Equal(t, tc.expected, report)
		})
	}
}

func TestMeshServiceDistributionToProviderReport(t *testing.T) {
	testCases := []struct {
		caseName string
		results  *ServiceDistributionResults
		expected types.ProviderReport
	}{
		{
			caseName: "service_distribution:istio=32,kuma=50,kong-mesh=50,traefik=20",
			results: &ServiceDistributionResults{
				TotalServices: 234,
				MeshDistribution: map[MeshKind]int{
					MeshKindIstio:    32,
					MeshKindKuma:     50,
					MeshKindKongMesh: 50,
					MeshKindTraefik:  20,
				},
			},
			expected: types.ProviderReport{
				"mdist": "all234,i32,k50,km50,t20",
			},
		},
		{
			caseName: "service_distribution:should return empty report for nil results",
			results:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			report := tc.results.ToProviderReport()
			require.Equal(t, tc.expected, report)
		})
	}
}
