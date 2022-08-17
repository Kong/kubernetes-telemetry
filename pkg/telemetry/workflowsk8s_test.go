package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
)

func TestWorkflowIdentifyPlatform(t *testing.T) {
	t.Run("basic construction fail for nil kubernetes.Interface", func(t *testing.T) {
		_, err := NewIdentifyPlatformWorkflow(nil)
		require.ErrorIs(t, err, ErrNilKubernetesInterfaceProvided)
	})

	t.Run("using fake client doesn't fail", func(t *testing.T) {
		kc := clientgo_fake.NewSimpleClientset()

		w, err := NewIdentifyPlatformWorkflow(kc)
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		require.EqualValues(t, provider.Report{
			provider.ClusterArchKey: fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			// Not really true but a reliable return value from client-go's fake client.
			provider.ClusterVersionKey:       "v0.0.0-master+$Format:%H$",
			provider.ClusterVersionSemverKey: "v0.0.0",
			provider.ClusterProviderKey:      provider.ClusterProviderUnknown,
		}, r)

		b, err := json.Marshal(r)
		require.NoError(t, err)
		fmt.Printf("%s\n", b)
	})
}

func TestWorkflowClusterState(t *testing.T) {
	t.Run("providing nil dynamic client fails", func(t *testing.T) {
		_, err := NewClusterStateWorkflow(nil, nil)
		require.ErrorIs(t, err, ErrNilDynClientProvided)
	})

	t.Run("properly reports cluster state", func(t *testing.T) {
		require.NoError(t, gatewayv1beta1.Install(scheme.Scheme))

		objs := []k8sruntime.Object{
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "kong-ingress-controller",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ingress-controller",
							Image: "kong/kubernetes-ingress-controller:2.4",
						},
					},
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace1",
					Name:      "srv",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "namespace2",
					Name:      "srv",
				},
			},
			&gatewayv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "gateway-1",
				},
			},

			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubeadm.alpha.kubernetes.io/cri-socket":                 "unix:///run/containerd/containerd.sock",
						"node.alpha.kubernetes.io/ttl":                           "0",
						"volumes.kubernetes.io/controller-managed-attach-detach": "true",
					},
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                                 "arm64",
						"beta.kubernetes.io/os":                                   "linux",
						"kubernetes.io/arch":                                      "arm64",
						"kubernetes.io/hostname":                                  "kong-control-plane",
						"kubernetes.io/os":                                        "linux",
						"node-role.kubernetes.io/control-plane":                   "",
						"node.kubernetes.io/exclude-from-external-load-balancers": "",
					},
					Name: "kong-control-plane",
				},
				Spec: corev1.NodeSpec{
					ProviderID: "gce://k8s/europe-north1-a/gke-cluster-user-default-pool-e1111111-aaii",
				},
			},
			&corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubeadm.alpha.kubernetes.io/cri-socket":                 "unix:///run/containerd/containerd.sock",
						"node.alpha.kubernetes.io/ttl":                           "0",
						"volumes.kubernetes.io/controller-managed-attach-detach": "true",
					},
					Labels: map[string]string{
						"beta.kubernetes.io/arch":                                 "arm64",
						"beta.kubernetes.io/os":                                   "linux",
						"kubernetes.io/arch":                                      "arm64",
						"kubernetes.io/hostname":                                  "worker-node-1",
						"kubernetes.io/os":                                        "linux",
						"node-role.kubernetes.io/control-plane":                   "",
						"node.kubernetes.io/exclude-from-external-load-balancers": "",
					},
					Name: "worker-node-1",
				},
			},
		}

		cl := ctrlclient_fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
		dynClient := dyn_fake.NewSimpleDynamicClientWithCustomListKinds(scheme.Scheme,
			map[schema.GroupVersionResource]string{
				{
					Group:    "gateway.networking.k8s.io",
					Version:  "v1beta1",
					Resource: "gateways",
				}: "GatewayList",
			},
			objs...,
		)

		w, err := NewClusterStateWorkflow(dynClient, cl.RESTMapper())
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		require.EqualValues(t, provider.Report{
			provider.NodeCountKey:    2,
			provider.PodCountKey:     1,
			provider.ServiceCountKey: 2,
			// TODO: Even though we added Gateway API's schema to schema.Scheme, gateway count provider
			// doesn't detect it properly due to:
			// https://github.com/kubernetes/kubernetes/pull/110053.
			// When that's addressed we should revisit this test.
			// provider.GatewayCountKey: 0,
		}, r)
	})
}
