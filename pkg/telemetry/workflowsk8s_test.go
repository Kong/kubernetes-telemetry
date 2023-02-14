package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
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
		require.EqualValues(t, types.ProviderReport{
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

		// With all this setup for Gateway API specific types we're able to get
		// List to work but it returns 0 elements because
		// https://github.com/kubernetes/client-go/blob/8ef4d7d4e87f691ab/testing/fixture.go#L258
		// holds an entry with GVR where Resource is "gatewaies".
		// Related: https://github.com/kubernetes/kubernetes/pull/110053.

		restMapper := meta.NewDefaultRESTMapper(nil)
		restMapper.AddSpecific(
			schema.GroupVersionKind{
				Group:   gatewayv1beta1.GroupVersion.Group,
				Version: gatewayv1beta1.GroupVersion.Version,
				Kind:    "Gateway",
			},
			schema.GroupVersionResource{
				Group:    gatewayv1beta1.GroupVersion.Group,
				Version:  gatewayv1beta1.GroupVersion.Version,
				Resource: "gateways",
			},
			schema.GroupVersionResource{
				Group:    gatewayv1beta1.GroupVersion.Group,
				Version:  gatewayv1beta1.GroupVersion.Version,
				Resource: "gateway",
			},
			meta.RESTScopeRoot,
		)

		cl := ctrlclient_fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithRuntimeObjects(objs...).
			WithRESTMapper(restMapper).
			Build()

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
		require.EqualValues(t, types.ProviderReport{
			provider.NodeCountKey:    2,
			provider.PodCountKey:     1,
			provider.ServiceCountKey: 2,
			// This should be equal to 1 but see above for comment explaining the issue.
			provider.GatewayCountKey: 0,
		}, r)
	})
}

func TestWorkflowMeshDetect(t *testing.T) {
	t.Run("providing nil client fails", func(t *testing.T) {
		_, err := NewMeshDetectWorkflow(nil, apitypes.NamespacedName{}, apitypes.NamespacedName{})
		require.ErrorIs(t, err, ErrNilControllerRuntimeClientProvided)
	})

	t.Run("properly reports cluster meshes", func(t *testing.T) {
		b := ctrlclient_fake.NewClientBuilder()
		b.WithObjects(
			// services.
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service1",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service2",
				},
			},
			&corev1.Service{
				// service with no available endpoints.
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service3",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service1",
				},
			},
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service2",
					Annotations: map[string]string{
						"mesh.traefik.io/traffic-type": "TCP",
					},
				},
			},
			// service with no endpoints.
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service3",
				},
			},
			// endpoints.
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service1",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns1", Name: "pod1"},
							},
						},
					},
				},
			},
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service2",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns1", Name: "pod2"},
							},
						},
					},
				},
			},
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service3",
				},
				// endpoints with no subsets.
			},
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service1",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod1"},
							},
						},
					},
				},
			},
			&corev1.Endpoints{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service2",
				},
				Subsets: []corev1.EndpointSubset{
					{
						Addresses: []corev1.EndpointAddress{
							{
								TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod2"},
							},
							{},
						},
					},
				},
			},
			// pods.
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "worker"},
						{Name: "istio-proxy"},
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "pod2",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "worker"},
						{Name: "kuma-sidecar"},
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "pod1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "worker"},
						{Name: "istio-proxy"},
						{Name: "linkerd-proxy"},
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "pod2",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "worker"},
					},
				},
			},
			// add pod
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "kong-ingress-controller",
					Annotations: map[string]string{
						"sidecar.istio.io/status":  "injected",
						"linkerd.io/proxy-version": "1.0.0",
						"kuma.io/sidecar-injected": "true",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ingress-controller",
							Image: "kong/kubernetes-ingress-controller:2.4",
						},
						// sidecars
						{Name: "istio-proxy"},
						{Name: "kuma-sidecar"},
						{
							Name:  "envoy",
							Image: "public.ecr.aws/appmesh/aws-appmesh-envoy:v1.22.2.0-prod",
						},
					},
					// init containers.
					InitContainers: []corev1.Container{
						{Name: "istio-init"},
					},
				},
			},
			// add kong-proxy service.
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "kong-proxy",
					Annotations: map[string]string{
						"mesh.traefik.io/traffic-type": "HTTP",
					},
				},
			},
			// add another namespace.
			&corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "kong-2",
					Annotations: map[string]string{
						"linkerd.io/inject": "enabled",
					},
					Labels: map[string]string{
						"appmesh.k8s.aws/sidecarInjectorWebhook": "enabled",
					},
				},
			},
			// add another pod and kong-proxy service.
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong-2",
					Name:      "kong-ingress-controller",
					Annotations: map[string]string{
						"linkerd.io/proxy-version":                   "1.0.0",
						"consul.hashicorp.com/connect-inject-status": "injected",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "ingress-controller",
							Image: "kong/kubernetes-ingress-controller:2.4",
						},
						// sidecars
						{Name: "linkerd-proxy"},
						{Name: "envoy-sidecar"},
					},
					// init containers.
					InitContainers: []corev1.Container{
						{Name: "linkerd-init"},
						{Name: "consul-connect-inject-init"},
					},
				},
			},
			// add a pod without a publishing service, and no injection.
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong-3",
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
					Namespace: "kong-2",
					Name:      "kong-proxy",
					Annotations: map[string]string{
						"mesh.traefik.io/traffic-type": "UDP",
					},
				},
			},
		)

		cl := b.Build()

		w, err := NewMeshDetectWorkflow(cl, apitypes.NamespacedName{
			Namespace: "kong",
			Name:      "kong-ingress-controller",
		}, apitypes.NamespacedName{})
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		require.EqualValues(t, types.ProviderReport{
			// TODO: the keys should be using consts but we can't place them in
			// provider package because that will cause an import cycle.
			// We could place them in mesh detect package but that would not be
			// consistent with other provider report keys.
			// Ideally we should revisit this at some point.
			"mdist": "all8,i2,k1,km1,l1,t1",
			"kinm":  "a3,i2,i3,i4,k2,k3,km2,km3,l2",
		}, r)
	})
}
