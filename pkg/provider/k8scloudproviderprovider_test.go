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
		// k3s
		{
			name: "k3s gets inferred from provider ID prefix",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "k3s://orbstack",
						},
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								"beta.kubernetes.io/arch":               "arm64",
								"beta.kubernetes.io/os":                 "linux",
								"kubernetes.io/arch":                    "arm64",
								"kubernetes.io/hostname":                "orbstack",
								"kubernetes.io/os":                      "linux",
								"node-role.kubernetes.io/control-plane": "",
							},
						},
					},
				)
			},
			expected: ClusterProviderK3S,
		},
		{
			name: "k3s gets inferred from annotations",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "k3s://orbstack",
						},
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"flannel.alpha.coreos.com/backend-data":                  "null",
								"flannel.alpha.coreos.com/backend-type":                  "host-gw",
								"flannel.alpha.coreos.com/backend-v6-data":               "null",
								"flannel.alpha.coreos.com/kube-subnet-manager":           "true",
								"flannel.alpha.coreos.com/public-ip":                     "198.19.249.2",
								"flannel.alpha.coreos.com/public-ipv6":                   "fd07:b51a:cc66::2",
								"k3s.io/hostname":                                        "orbstack",
								"k3s.io/internal-ip":                                     "198.19.249.2,fd07:b51a:cc66::2",
								"k3s.io/node-args":                                       `["server","--disable","metrics-server,traefik,coredns","--https-listen-port","26443","--lb-server-port","26444","--docker","--container-runtime-endpoint","/var/run/docker.sock","--protect-kernel-defaults","--flannel-backend","host-gw","--cluster-cidr","192.168.194.0/25,fd07:b51a:cc66:a::/72","--service-cidr","192.168.194.128/25,fd07:b51a:cc66:a:8000::/112","--kube-controller-manager-arg","node-cidr-mask-size-ipv4=25","--kube-controller-manager-arg","node-cidr-mask-size-ipv6=72","--write-kubeconfig","/run/kubeconfig.yml"]`,
								"k3s.io/node-config-hash":                                "sjrlvsvk5fwsmiduwp6tm23ofcbjl3t4qk4p3sozvzjpzaefar2a====",
								"k3s.io/node-env":                                        `{"k3s_data_dir"":/var/lib/rancher/k3s/data/8f922a49dfab9ed75d85cba9b81b9ac7f20a633da393551392030e41ea4d569b"}`,
								"node.alpha.kubernetes.io/ttl":                           "0",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"beta.kubernetes.io/arch":               "arm64",
								"beta.kubernetes.io/os":                 "linux",
								"kubernetes.io/arch":                    "arm64",
								"kubernetes.io/hostname":                "orbstack",
								"kubernetes.io/os":                      "linux",
								"node-role.kubernetes.io/control-plane": "",
							},
						},
					},
				)
			},
			expected: ClusterProviderK3S,
		},
		{
			name: "Azure node inferred from node labels",
			clientFunc: func() *clientgo_fake.Clientset {
				return clientgo_fake.NewSimpleClientset(
					&corev1.Node{
						Spec: corev1.NodeSpec{
							ProviderID: "azure:///subscriptions/1111111-422b-9a21-fd111111111/resourceGroups/mc_asd_asd_eucentral/providers/Microsoft.Compute/virtualMachineScaleSets/aks-nodepool1-1111-vmss/virtualMachines/0",
						},
						ObjectMeta: metav1.ObjectMeta{
							Annotations: map[string]string{
								"csi.volume.kubernetes.io/nodeid":                        `{"disk.csi.azure.com":"aks-nodepool1-20479306-vmss000000","file.csi.azure.com":"aks-nodepool1-20479306-vmss000000"}`,
								"node.alpha.kubernetes.io/ttl":                           "0",
								"volumes.kubernetes.io/controller-managed-attach-detach": "true",
							},
							Labels: map[string]string{
								"agentpool":                                               "nodepool1",
								"beta.kubernetes.io/arch":                                 "amd64",
								"beta.kubernetes.io/instance-type":                        "Standard_DS2_v2",
								"beta.kubernetes.io/os":                                   "linux",
								"failure-domain.beta.kubernetes.io/region":                "polandcentral",
								"failure-domain.beta.kubernetes.io/zone":                  "0",
								"kubernetes.azure.com/agentpool":                          "nodepool1",
								"kubernetes.azure.com/cluster":                            "MC_asd_asd_eucentral",
								"kubernetes.azure.com/consolidated-additional-properties": "6ea4bf17-2fb7-11ef-ac0b-fd1111111111",
								"kubernetes.azure.com/kubelet-identity-client-id":         "42e89310-1d30-43fd-8888-fd1111111111",
								"kubernetes.azure.com/mode":                               "system",
								"kubernetes.azure.com/node-image-version":                 "AKSUbuntu-2204gen2containerd-202405.27.0",
								"kubernetes.azure.com/nodepool-type":                      "VirtualMachineScaleSets",
								"kubernetes.azure.com/os-sku":                             "Ubuntu",
								"kubernetes.azure.com/role":                               "agent",
								"kubernetes.azure.com/storageprofile":                     "managed",
								"kubernetes.azure.com/storagetier":                        "Premium_LRS",
								"kubernetes.io/arch":                                      "amd64",
								"kubernetes.io/hostname":                                  "aks-nodepool1-111111111-vmss000000",
								"kubernetes.io/os":                                        "linux",
								"kubernetes.io/role":                                      "agent",
								"node-role.kubernetes.io/agent":                           "",
								"node.kubernetes.io/instance-type":                        "Standard_DS2_v2",
								"storageprofile":                                          "managed",
								"storagetier":                                             "Premium_LRS",
								"topology.disk.csi.azure.com/zone":                        "",
								"topology.kubernetes.io/region":                           "polandcentral",
								"topology.kubernetes.io/zone":                             "0",
							},
						},
					},
				)
			},
			expected: ClusterProviderAzure,
		},
		{
			name: "Rancher node inferred from version",
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
					Minor:        "27+",
					GitVersion:   "v1.27.8+rke2r1",
					GitCommit:    "cc6a1b4915a99f49f5510ef0667f94b9ca832a8a",
					GitTreeState: "clean",
					BuildDate:    "2024-03-09T18:24:04Z",
					GoVersion:    "go1.21.15",
					Compiler:     "gc",
					Platform:     "linux/amd64",
				}

				return kc
			},
			expected: ClusterProviderRKE2,
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
