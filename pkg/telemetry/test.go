package telemetry

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// Scheme returns a Kubernetes scheme with the necessary types registered for testing.
func Scheme(t testing.TB) *runtime.Scheme {
	s := runtime.NewScheme()

	require.NoError(t, metav1.AddMetaToScheme(s))
	require.NoError(t, corev1.AddToScheme(s))
	require.NoError(t, netv1.AddToScheme(s))

	require.NoError(t, gatewayv1.Install(s))
	require.NoError(t, gatewayv1beta1.Install(s))
	require.NoError(t, gatewayv1alpha2.Install(s))

	return s
}

// toPartialObjectMetadata converts typed Kubernetes objects into
// *metav1.PartialObjectMetadata so they can be used with the metadata fake client.
func toPartialObjectMetadata(s *runtime.Scheme, objs ...runtime.Object) []runtime.Object {
	out := make([]runtime.Object, 0, len(objs))
	for _, obj := range objs {
		gvks, _, err := s.ObjectKinds(obj)
		if err != nil || len(gvks) == 0 {
			continue
		}
		accessor, err := meta.Accessor(obj)
		if err != nil {
			continue
		}
		out = append(out, &metav1.PartialObjectMetadata{
			TypeMeta: metav1.TypeMeta{
				APIVersion: gvks[0].GroupVersion().String(),
				Kind:       gvks[0].Kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      accessor.GetName(),
				Namespace: accessor.GetNamespace(),
			},
		})
	}
	return out
}
