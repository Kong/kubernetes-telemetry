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
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"

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
		_, err := NewClusterStateWorkflow(nil)
		require.ErrorIs(t, err, ErrNilDynClientProvided)
	})

	t.Run("properly reports cluster state", func(t *testing.T) {
		dynClient := dyn_fake.NewSimpleDynamicClient(scheme.Scheme,
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
		)

		w, err := NewClusterStateWorkflow(dynClient)
		require.NoError(t, err)
		require.NotNil(t, w)

		r, err := w.Execute(context.Background())
		require.NoError(t, err)
		require.NotNil(t, r)
		require.EqualValues(t, provider.Report{
			provider.PodCountKey:     1,
			provider.ServiceCountKey: 2,
		}, r)
	})
}
