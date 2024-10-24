package meshdetect

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestDetectMeshDeployment(t *testing.T) {
	testScheme := runtime.NewScheme()
	err := corev1.AddToScheme(testScheme)
	require.NoErrorf(t, err, "should add corev1 to scheme successfully")

	b := fake.NewClientBuilder().
		WithScheme(testScheme).
		WithIndex(&corev1.Service{}, "metadata.name", func(object client.Object) []string {
			return []string{object.GetNamespace(), object.GetName()}
		})
	b.WithObjects(
		// add services.
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "istio-system",
				Name:      "istiod",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kuma-sys",
				Name:      "kong-mesh-control-plane",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "traefik-mesh-controller",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "consul-mesh",
				Name:      "consul-server",
			},
		},
	)

	c := b.Build()
	d := &Detector{
		Client: c,
		Pod: apitypes.NamespacedName{
			Name:      "kic-1",
			Namespace: "kong",
		},
	}

	res := d.DetectMeshDeployment(context.Background())
	expected := map[MeshKind]*DeploymentResults{
		MeshKindIstio: {
			ServiceExists: true,
		},
		MeshKindLinkerd: {
			ServiceExists: false,
		},
		MeshKindKuma: {
			ServiceExists: false,
		},
		MeshKindKongMesh: {
			ServiceExists: true,
		},
		MeshKindConsul: {
			ServiceExists: true,
		},
		MeshKindTraefik: {
			ServiceExists: true,
		},
		MeshKindAWSAppMesh: {
			ServiceExists: false,
		},
	}

	for _, meshKind := range MeshesToDetect {
		t.Run(string(meshKind), func(t *testing.T) {
			require.Equalf(t, expected[meshKind], res[meshKind], "detection result should be the same for mesh %s", meshKind)
		})
	}
}

func TestDetectRunUnder(t *testing.T) {
	b := fake.NewClientBuilder()
	b.WithObjects(
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
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kong-3",
				Name:      "kong-proxy",
			},
		},
	)

	testCases := []struct {
		caseName        string
		pod             apitypes.NamespacedName
		expectedResults map[MeshKind]*RunUnderResults
	}{
		{
			caseName: "injected-istio,kuma,traefik,aws;annotation-linkerd",
			pod: apitypes.NamespacedName{
				Name:      "kong-ingress-controller",
				Namespace: "kong",
			},
			expectedResults: map[MeshKind]*RunUnderResults{
				MeshKindIstio: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    true,
				},
				MeshKindLinkerd: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: false,
					InitContainerInjected:    false,
				},
				MeshKindKuma: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    false,
				},
				MeshKindKongMesh: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    false,
				},
				MeshKindConsul: {
					// all false
				},
				MeshKindTraefik: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: false,
					InitContainerInjected:    false,
				},
				MeshKindAWSAppMesh: {
					PodOrServiceAnnotation:   false,
					SidecarContainerInjected: true,
					InitContainerInjected:    false,
				},
			},
		},
		{
			caseName: "injected-linkerd,consul",
			pod: apitypes.NamespacedName{
				Name:      "kong-ingress-controller",
				Namespace: "kong-2",
			},
			expectedResults: map[MeshKind]*RunUnderResults{
				MeshKindIstio: {
					// all false
				},
				MeshKindLinkerd: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    true,
				},
				MeshKindKuma: {
					// all false
				},
				MeshKindKongMesh: {
					// all false
				},
				MeshKindConsul: {
					PodOrServiceAnnotation:   true,
					SidecarContainerInjected: true,
					InitContainerInjected:    true,
				},
				MeshKindTraefik: {
					// all false
				},
				MeshKindAWSAppMesh: {
					PodOrServiceAnnotation:   false,
					SidecarContainerInjected: false,
					InitContainerInjected:    false,
				},
			},
		},
		{
			caseName: "nothing injected",
			pod: apitypes.NamespacedName{
				Name:      "kong-ingress-controller",
				Namespace: "kong-3",
			},
			expectedResults: map[MeshKind]*RunUnderResults{
				// all mesh kinds -> all false
				MeshKindIstio:      {},
				MeshKindLinkerd:    {},
				MeshKindKuma:       {},
				MeshKindKongMesh:   {},
				MeshKindConsul:     {},
				MeshKindTraefik:    {},
				MeshKindAWSAppMesh: {},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.caseName, func(t *testing.T) {
			d := &Detector{
				Client: b.Build(),
				Pod:    tc.pod,
				PublishService: apitypes.NamespacedName{
					Namespace: tc.pod.Namespace,
					Name:      "kong-proxy",
				},
			}
			res, err := d.DetectRunUnder(context.Background())
			require.NoError(t, err)
			for _, meshKind := range MeshesToDetect {
				require.Equalf(t, tc.expectedResults[meshKind], res[meshKind],
					"test case %s: detection result should be same for mesh %s", tc.caseName, meshKind)
			}
		})
	}
}

func TestDetectServiceDistribution(t *testing.T) {
	b := fake.NewClientBuilder()
	// Add services/endpoints/pods.
	b.WithObjects(
		// Services.
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
	)

	c := b.Build()
	d := &Detector{
		Client: c,
	}

	const expectedTotal = 7
	expected := map[MeshKind]int{
		MeshKindIstio:    2,
		MeshKindLinkerd:  2,
		MeshKindKuma:     1,
		MeshKindKongMesh: 1,
		MeshKindTraefik:  1,
	}

	res, err := d.DetectServiceDistribution(context.Background())
	require.NoErrorf(t, err, "should not return error in detecting service distribution")
	require.Equalf(t, expectedTotal, res.TotalServices, "total number of services should be the same")
	for _, meshKind := range MeshesToDetect {
		require.Equalf(t, expected[meshKind], res.MeshDistribution[meshKind],
			"service within mesh %s should be same as expected", meshKind)
	}
}
