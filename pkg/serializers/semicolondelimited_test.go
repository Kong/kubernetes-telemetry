package serializers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestSemicolonDelimited(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		s := NewSemicolonDelimited("kic-ping")

		out, err := s.Serialize(types.Report{
			"cluster-state": provider.Report{
				"k8s_pods_count":     1,
				"k8s_services_count": 2,
			},
			"identify-platform": provider.Report{
				"k8s_arch":     "linux/arm64",
				"k8sv":         "v1.2.3-gke-a1fdc32f",
				"k8sv_semver":  "v1.2.3",
				"k8s_provider": provider.ClusterProviderGKE,
			},
		})

		require.NoError(t, err)
		assert.EqualValues(t, "<14>signal=kic-ping;k8s_arch=linux/arm64;k8s_provider=GKE;k8sv=v1.2.3-gke-a1fdc32f;k8sv_semver=v1.2.3;k8s_pods_count=1;k8s_services_count=2;\n", string(out))
	})
}
