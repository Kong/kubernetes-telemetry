package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestClusterProvider(t *testing.T) {
	testcases := []struct {
		name       string
		clientFunc func() *clientgo_fake.Clientset
		expected   ClusterProvider
	}{
		{
			name: "no objects in the cluster return unknown cluster provider",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset()
			},
			expected: ClusterProviderUnknown,
		},
		// GKE
		{
			name: "gke node (without correctly set provider ID, labels nor annotations) makes the provider return gke based on version string",
			clientFunc: func() *clientgo_fake.Clientset {
				kc := clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl": "0",
							},
						},
					},
				)

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
			expected: ClusterProviderGKE,
		},
		{
			name: "gke node makes the provider return gke provider as a result",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "gce://k8s-playground/europe-north1-a/gke-cluster-user-default-pool-e123123123-aaii",
						},
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"container.googleapis.com/instance_id":                   "8474243960111131104",
								"csi.volume.kubernetes.io/nodeid":                        `{"pd.csi.storage.gke.io":"projects/k8s-playground/zones/europe-north1-a/instances/gke-cluster-user-default-pool-123123123-cdht"}`,
								"node.alpha.kubernetes.io/ttl":                           "0",
								"node.gke.io/last-applied-node-labels":                   "cloud.google.com/gke-boot-disk=pd-standard,cloud.google.com/gke-container-runtime=containerd,cloud.google.com/gke-cpu-scaling-level=2,cloud.google.com/gke-max-pods-per-node=110,cloud.google.com/gke-nodepool=default-pool,cloud.google.com/gke-os-distribution=cos,cloud.google.com/gke-spot=true,cloud.google.com/machine-family=e2,cloud.google.com/private-node=false",
								"node.gke.io/last-applied-node-taints":                   "",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"beta.kubernetes.io/arch":                  "amd64",
								"beta.kubernetes.io/instance-type":         "e2-medium",
								"beta.kubernetes.io/os":                    "linux",
								"cloud.google.com/gke-boot-disk":           "pd-standard",
								"cloud.google.com/gke-container-runtime":   "containerd",
								"cloud.google.com/gke-cpu-scaling-level":   "2",
								"cloud.google.com/gke-max-pods-per-node":   "110",
								"cloud.google.com/gke-nodepool":            "default-pool",
								"cloud.google.com/gke-os-distribution":     "cos",
								"cloud.google.com/gke-spot":                "true",
								"cloud.google.com/machine-family":          "e2",
								"cloud.google.com/private-node":            "false",
								"failure-domain.beta.kubernetes.io/region": "europe-north1",
								"failure-domain.beta.kubernetes.io/zone":   "europe-north1-a",
								"kubernetes.io/arch":                       "amd64",
								"kubernetes.io/hostname":                   "gke-cluster-user-default-pool-123123123-cdht",
								"kubernetes.io/os":                         "linux",
								"node.kubernetes.io/instance-type":         "e2-medium",
								"topology.gke.io/zone":                     "europe-north1-a",
								"topology.kubernetes.io/region":            "europe-north1",
								"topology.kubernetes.io/zone":              "europe-north1-a",
							},
						},
					},
				)
			},
			expected: ClusterProviderGKE,
		},
		{
			name: "gke node (without gco:// provider ID) makes the provider return gke based on labels and annotations",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"container.googleapis.com/instance_id":                   "8474243960111131104",
								"csi.volume.kubernetes.io/nodeid":                        `{"pd.csi.storage.gke.io":"projects/k8s-playground/zones/europe-north1-a/instances/gke-cluster-user-default-pool-123123123-cdht"}`,
								"node.alpha.kubernetes.io/ttl":                           "0",
								"node.gke.io/last-applied-node-labels":                   "cloud.google.com/gke-boot-disk=pd-standard,cloud.google.com/gke-container-runtime=containerd,cloud.google.com/gke-cpu-scaling-level=2,cloud.google.com/gke-max-pods-per-node=110,cloud.google.com/gke-nodepool=default-pool,cloud.google.com/gke-os-distribution=cos,cloud.google.com/gke-spot=true,cloud.google.com/machine-family=e2,cloud.google.com/private-node=false",
								"node.gke.io/last-applied-node-taints":                   "",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"beta.kubernetes.io/arch":                  "amd64",
								"beta.kubernetes.io/instance-type":         "e2-medium",
								"beta.kubernetes.io/os":                    "linux",
								"cloud.google.com/gke-boot-disk":           "pd-standard",
								"cloud.google.com/gke-container-runtime":   "containerd",
								"cloud.google.com/gke-cpu-scaling-level":   "2",
								"cloud.google.com/gke-max-pods-per-node":   "110",
								"cloud.google.com/gke-nodepool":            "default-pool",
								"cloud.google.com/gke-os-distribution":     "cos",
								"cloud.google.com/gke-spot":                "true",
								"cloud.google.com/machine-family":          "e2",
								"cloud.google.com/private-node":            "false",
								"failure-domain.beta.kubernetes.io/region": "europe-north1",
								"failure-domain.beta.kubernetes.io/zone":   "europe-north1-a",
								"kubernetes.io/arch":                       "amd64",
								"kubernetes.io/hostname":                   "gke-cluster-user-default-pool-123123123-cdht",
								"kubernetes.io/os":                         "linux",
								"node.kubernetes.io/instance-type":         "e2-medium",
								"topology.gke.io/zone":                     "europe-north1-a",
								"topology.kubernetes.io/region":            "europe-north1",
								"topology.kubernetes.io/zone":              "europe-north1-a",
							},
						},
					},
				)
			},
			expected: ClusterProviderGKE,
		},
		// AWS
		{
			name: "aws node makes the provider return aws",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "aws:///eu-west-1b/i-0fa11111111111111",
						},
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl":                           "0",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"alpha.eksctl.io/cluster-name":             "spot",
								"alpha.eksctl.io/instance-id":              "i-0fac1111111111111",
								"alpha.eksctl.io/nodegroup-name":           "ng-capacity-optimized-b",
								"beta.kubernetes.io/arch":                  "amd64",
								"beta.kubernetes.io/instance-type":         "r5n.large",
								"beta.kubernetes.io/os":                    "linux",
								"failure-domain.beta.kubernetes.io/region": "eu-west-1",
								"failure-domain.beta.kubernetes.io/zone":   "eu-west-1b",
								"kubernetes.io/arch":                       "amd64",
								"kubernetes.io/hostname":                   "ip-192-168-21-1.eu-west-1.compute.internal",
								"kubernetes.io/os":                         "linux",
								"lifecycle":                                "Ec2Spot",
								"node-lifecycle":                           "spot",
								"node.kubernetes.io/instance-type":         "r5n.large",
								"topology.kubernetes.io/region":            "eu-west-1",
								"topology.kubernetes.io/zone":              "eu-west-1b",
							},
						},
					},
				)
			},
			expected: ClusterProviderAWS,
		},
		{
			name: "aws node (without correctly set provider ID) makes the provider return aws based on labels",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl":                           "0",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"alpha.eksctl.io/cluster-name":             "spot",
								"alpha.eksctl.io/instance-id":              "i-0fac1111111111111",
								"alpha.eksctl.io/nodegroup-name":           "ng-capacity-optimized-b",
								"beta.kubernetes.io/arch":                  "amd64",
								"beta.kubernetes.io/instance-type":         "r5n.large",
								"beta.kubernetes.io/os":                    "linux",
								"failure-domain.beta.kubernetes.io/region": "eu-west-1",
								"failure-domain.beta.kubernetes.io/zone":   "eu-west-1b",
								"kubernetes.io/arch":                       "amd64",
								"kubernetes.io/hostname":                   "ip-192-168-21-1.eu-west-1.compute.internal",
								"kubernetes.io/os":                         "linux",
								"lifecycle":                                "Ec2Spot",
								"node-lifecycle":                           "spot",
								"node.kubernetes.io/instance-type":         "r5n.large",
								"topology.kubernetes.io/region":            "eu-west-1",
								"topology.kubernetes.io/zone":              "eu-west-1b",
							},
						},
					},
				)
			},
			expected: ClusterProviderAWS,
		},
		{
			name: "aws node x(without correctly set provider ID, labels nor annotations) makes the provider return aws based on version string",
			clientFunc: func() *clientgo_fake.Clientset {
				kc := clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"node.alpha.kubernetes.io/ttl":                           "0",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
						},
					},
				)

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
			expected: ClusterProviderAWS,
		},
		// kind
		{
			name: "kind gets inferred from provider ID prefix",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "kind://docker/kong/kong-control-plane",
						},
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"beta.kubernetes.io/arch":                                 "arm64",
								"beta.kubernetes.io/os":                                   "linux",
								"kubernetes.io/arch":                                      "arm64",
								"kubernetes.io/hostname":                                  "kong-control-plane",
								"kubernetes.io/os":                                        "linux",
								"node-role.kubernetes.io/control-plane":                   "",
								"node.kubernetes.io/exclude-from-external-load-balancers": "",
							},
						},
					},
				)
			},
			expected: ClusterProviderKubernetesInDocker,
		},
	}

	for _, tc := range testcases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			p, err := NewK8sClusterProviderProvider(tc.name, tc.clientFunc())
			require.NoError(t, err)

			r, err := p.Provide(context.Background())
			require.NoError(t, err)
			require.EqualValues(t, types.ProviderReport{
				ClusterProviderKey: tc.expected,
			}, r)
		})
	}
}
