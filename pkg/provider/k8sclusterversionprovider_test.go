package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestClusterVersion(t *testing.T) {
	testcases := []struct {
		name       string
		clientFunc func() *clientgo_fake.Clientset
		expected   types.ProviderReport
	}{
		{
			name: "undecodable git version from /version API returns the major and minor version concatenated string",
			clientFunc: func() *clientgo_fake.Clientset {
				kc := clientgo_fake.NewSimpleClientset()

				d, ok := kc.Discovery().(*fakediscovery.FakeDiscovery)
				require.True(t, ok)
				d.FakedServerVersion = &version.Info{
					Major:        "1",
					Minor:        "24",
					GitVersion:   "v1-custom",
					GitCommit:    "cc6a1b4915a99f49f5510ef0667f94b9ca832a8a",
					GitTreeState: "clean",
					BuildDate:    "2022-06-09T18:24:04Z",
					GoVersion:    "go1.16.15",
					Compiler:     "gc",
					Platform:     "linux/amd64",
				}

				return kc
			},
			expected: types.ProviderReport{
				ClusterVersionKey:       "v1-custom",
				ClusterVersionSemverKey: "v1.24",
			},
		},
		// GKE
		{
			name: "gke versioning scheme is decoded properly",
			clientFunc: func() *clientgo_fake.Clientset {
				kc := clientgo_fake.NewSimpleClientset()

				d, ok := kc.Discovery().(*fakediscovery.FakeDiscovery)
				require.True(t, ok)
				d.FakedServerVersion = &version.Info{
					Major:        "1",
					Minor:        "24",
					GitVersion:   "v1.24.1-gke.1400",
					GitCommit:    "206efe6f8106824435f9e408af6f49e61b30ae54",
					GitTreeState: "clean",
					BuildDate:    "2022-06-13T19:52:07Z",
					GoVersion:    "go1.18.2b7",
					Compiler:     "gc",
					Platform:     "linux/amd64",
				}

				return kc
			},
			expected: types.ProviderReport{
				ClusterVersionKey:       "v1.24.1-gke.1400",
				ClusterVersionSemverKey: "v1.24.1",
			},
		},
		// AWS
		{
			name: "aws versioning scheme is decoded properly",
			clientFunc: func() *clientgo_fake.Clientset {
				kc := clientgo_fake.NewSimpleClientset()

				d, ok := kc.Discovery().(*fakediscovery.FakeDiscovery)
				require.True(t, ok)
				d.FakedServerVersion = &version.Info{
					Major:        "1",
					Minor:        "22+",
					GitVersion:   "v1.22.10-eks-84b4fe6",
					GitCommit:    "cc6a1b4915a99f49f5510ef0667f94b9ca832a8a",
					GitTreeState: "clean",
					BuildDate:    "2022-06-09T18:24:04Z",
					GoVersion:    "go1.16.15",
					Compiler:     "gc",
					Platform:     "linux/amd64",
				}

				return kc
			},
			expected: types.ProviderReport{
				ClusterVersionKey:       "v1.22.10-eks-84b4fe6",
				ClusterVersionSemverKey: "v1.22.10",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewK8sClusterVersionProvider(tc.name, tc.clientFunc())
			require.NoError(t, err)

			r, err := p.Provide(context.Background())
			require.NoError(t, err)
			require.EqualValues(t, tc.expected, r)
		})
	}
}
