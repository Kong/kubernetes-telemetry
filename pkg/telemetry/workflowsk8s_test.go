package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apitypes "k8s.io/apimachinery/pkg/types"
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	exampleOpenShiftVersion = "4.13.0"
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

		// this test does not provide an OpenShift operator Pod to the fake client, so the report should omit this key
		_, hasOpenShiftKey := r[provider.OpenShiftVersionKey]
		require.False(t, hasOpenShiftKey)

		b, err := json.Marshal(r)
		require.NoError(t, err)
		fmt.Printf("%s\n", b)
	})

	t.Run("OpenShift version detection returns version when OpenShift operators are present", func(t *testing.T) {
		kc := clientgo_fake.NewSimpleClientset(generateOpenShiftObjects()...)

		w, err := NewIdentifyPlatformWorkflow(kc)
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		require.EqualValues(t, exampleOpenShiftVersion, r[provider.OpenShiftVersionKey])

		b, err := json.Marshal(r)
		require.NoError(t, err)
		fmt.Printf("%s\n", b)
	})
}

func generateOpenShiftObjects() []k8sruntime.Object {
	return []k8sruntime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: provider.OpenShiftVersionPodNamespace,
			},
			Spec: corev1.NamespaceSpec{},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: provider.OpenShiftVersionPodNamespace,
				Name:      provider.OpenShiftVersionPodApp + "-85c4c6dbb7-zbrkm",
				Labels: map[string]string{
					"app": provider.OpenShiftVersionPodApp,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "worker",
						Env: []corev1.EnvVar{
							{
								Name:  provider.ImageVersionVariable,
								Value: exampleOpenShiftVersion,
							},
						},
					},
				},
			},
		},
	}
}

func TestWorkflowClusterState(t *testing.T) {
	t.Run("providing nil dynamic client fails", func(t *testing.T) {
		_, err := NewClusterStateWorkflow(nil, nil)
		require.ErrorIs(t, err, ErrNilDynClientProvided)
	})

	t.Run("properly reports cluster state", func(t *testing.T) {
		require.NoError(t, gatewayv1.Install(scheme.Scheme))
		require.NoError(t, gatewayv1beta1.Install(scheme.Scheme))
		require.NoError(t, gatewayv1alpha2.Install(scheme.Scheme))

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
			&gatewayv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "gatewayclass-1",
				},
			},
			&gatewayv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "gateway-1",
				},
			},
			&gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "httproute-1",
				},
			},
			&gatewayv1beta1.ReferenceGrant{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "referencegrant-1",
				},
			},
			&gatewayv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "grpcroute-1",
				},
			},
			&gatewayv1alpha2.TCPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "tcproute-1",
				},
			},
			&gatewayv1alpha2.UDPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "udproute-1",
				},
			},
			&gatewayv1alpha2.TLSRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "tlsroute-1",
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

		// With all this setup for Gateway API types we're able to get them
		// to work but for Kind "Gateway" it returns 0 elements because
		// https://github.com/kubernetes/client-go/blob/8ef4d7d4e87f691ab/testing/fixture.go#L258
		// holds an entry with GVR where Resource is "gatewaies".
		// Related: https://github.com/kubernetes/kubernetes/pull/110053.

		// All Gateway API types have to be explicitly added to the RESTMapper,
		// to make it work with the client-go fake client.
		restMapper := meta.NewDefaultRESTMapper(nil)
		as := func(gv metav1.GroupVersion, kind, resourcePlural string) {
			restMapper.AddSpecific(
				schema.GroupVersionKind{
					Group:   gv.Group,
					Version: gv.Version,
					Kind:    kind,
				},
				schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: resourcePlural,
				},
				schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: strings.ToLower(kind),
				},
				meta.RESTScopeRoot,
			)
		}
		as(gatewayv1.GroupVersion, "GatewayClass", "gatewayclasses")
		as(gatewayv1.GroupVersion, "Gateway", "gateways")
		as(gatewayv1.GroupVersion, "HTTPRoute", "httproutes")
		as(gatewayv1.GroupVersion, "GRPCRoute", "grpcroutes")
		as(gatewayv1beta1.GroupVersion, "ReferenceGrant", "referencegrants")
		as(gatewayv1alpha2.GroupVersion, "TCPRoute", "tcproutes")
		as(gatewayv1alpha2.GroupVersion, "UDPRoute", "udproutes")
		as(gatewayv1alpha2.GroupVersion, "TLSRoute", "tlsroutes")

		cl := ctrlclient_fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithRuntimeObjects(objs...).
			WithRESTMapper(restMapper).
			Build()

		// Hack for Kind "Gateway" to work.
		dynClient := dyn_fake.NewSimpleDynamicClientWithCustomListKinds(
			scheme.Scheme,
			map[schema.GroupVersionResource]string{
				{
					Group:    "gateway.networking.k8s.io",
					Version:  "v1",
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
			// core v1
			provider.NodeCountKey:    2,
			provider.PodCountKey:     1,
			provider.ServiceCountKey: 2,
			// gateway.networking.k8s.io v1
			provider.GRPCRouteCountKey:    1,
			provider.HTTPRouteCountKey:    1,
			provider.GatewayClassCountKey: 1,
			provider.GatewayCountKey:      0, // This should be equal to 1 but see above for comment explaining the issue.
			// gateway.networking.k8s.io v1beta1
			provider.ReferenceGrantCountKey: 1,
			// gateway.networking.k8s.io v1alpha2
			provider.TCPRouteCountKey: 1,
			provider.UDPRouteCountKey: 1,
			provider.TLSRouteCountKey: 1,
		}, r)
	})

	t.Run("properly reports cluster state without GW API objects when their CRDs are missing", func(t *testing.T) {
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

		cl := ctrlclient_fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithRuntimeObjects(objs...).
			Build()

		// Hack for Kind "Gateway" to work.
		dynClient := dyn_fake.NewSimpleDynamicClient(
			scheme.Scheme,
			objs...,
		)

		w, err := NewClusterStateWorkflow(dynClient, cl.RESTMapper())
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		// This technically wouldn't fail with missing error check against
		// discovery.ErrGroupDiscoveryFailed type in NewClusterStateWorkflow()
		// because test code uses apimachinery restmapper [1] while production
		// code when running the operator uses the controller runtime mapper [2].
		//
		// [1]: https://github.com/kubernetes/apimachinery/blob/16053f78e9258e6c227f5e704d35c67e61868d12/pkg/api/meta/restmapper.go#L360-L362
		// [2]: https://github.com/kubernetes-sigs/controller-runtime/blob/e54088c8c7da82111b4508bdaf189c45d1344f00/pkg/client/apiutil/restmapper.go#L67-L69
		require.EqualValues(t, types.ProviderReport{
			// core v1
			provider.NodeCountKey:    2,
			provider.PodCountKey:     1,
			provider.ServiceCountKey: 2,
			// gateway.networking.k8s.io
			// No Gateway API related objects are reported and also no error is returned.
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
			// Service with no EndpointSlices.
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service3",
				},
			},
			// Service with multiple EndpointSlices.
			&corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service4",
				},
			},
			// EndpointSlices for Pods.
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service1-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service1",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns1", Name: "pod1"},
					},
				},
			},
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service2-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service2",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns1", Name: "pod2"},
					},
				},
			},
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "service3-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service3",
					},
				},
				// EndpointSlice with no endpoints.
			},
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service1-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service1",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod1"},
					},
				},
			},
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service2-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service2",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod2"},
					},
					{},
				},
			},
			// Two EndpointSlices for the same service.
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service3-1",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service3",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod3-1"},
					},
				},
			},
			&discoveryv1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "service3-2",
					Labels: map[string]string{
						discoveryv1.LabelServiceName: "service3",
					},
				},
				Endpoints: []discoveryv1.Endpoint{
					{
						TargetRef: &corev1.ObjectReference{Kind: "Pod", Namespace: "ns2", Name: "pod3-2"},
					},
				},
			},
			// Pods referenced by EndpointSlices.
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
			// One Pod has service mesh sidecar, the other doesn't.
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "pod3-1",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "worker"},
						{Name: "linkerd-proxy"},
					},
				},
			},
			&corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns2",
					Name:      "pod3-2",
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
			"mdist": "all9,i2,k1,km1,l2,t1",
			"kinm":  "a3,i2,i3,i4,k2,k3,km2,km3,l2",
		}, r)
	})
}
