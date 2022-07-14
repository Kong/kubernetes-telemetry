package telemetry

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/Kong/kubernetes-telemetry/pkg/provider"
)

func TestManagerStartStopDoesntFail(t *testing.T) {
	m, err := NewManager(OptManagerLogger(logr.Discard()))
	require.NoError(t, err)
	require.NoError(t, m.Start())
	m.Stop()
}

func TestManagerBasicLogicWorks(t *testing.T) {
	m, err := NewManager(
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
		{
			p, err := provider.NewFixedValueProvider("constant1", provider.Report{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", provider.Report{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	ch := make(chan Report)
	require.NoError(t, m.AddConsumer(ch))
	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)
	require.ErrorIs(t, m.AddConsumer(make(chan<- Report)),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	report := <-ch
	m.Stop()
	require.EqualValues(t, Report{
		"basic1": provider.Report{
			"constant1": "value1",
			"constant2": "value2",
		},
	}, report)
}

func TestManagerWithMultilpleWorkflows(t *testing.T) {
	m, err := NewManager(
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
		{
			p, err := provider.NewFixedValueProvider("constant1", provider.Report{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", provider.Report{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}
	{
		w := NewWorkflow("basic2")
		{
			p, err := provider.NewFixedValueProvider("constant1", provider.Report{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", provider.Report{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	ch := make(chan Report)
	require.NoError(t, m.AddConsumer(ch))
	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)
	require.ErrorIs(t, m.AddConsumer(make(chan<- Report)),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	report := <-ch
	m.Stop()
	require.EqualValues(t, Report{
		"basic1": provider.Report{
			"constant1": "value1",
			"constant2": "value2",
		},
		"basic2": provider.Report{
			"constant1": "value1",
			"constant2": "value2",
		},
	}, report)
}

func TestManagerWithCatalogWorkflows(t *testing.T) {
	t.Run("identify platform and cluster state", func(t *testing.T) {
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
		clusterStateWorkflow, err := NewClusterStateWorkflow(dynClient)
		require.NoError(t, err)
		require.NotNil(t, clusterStateWorkflow)

		kc := clientgo_fake.NewSimpleClientset()
		identifyPlatformWorkflow, err := NewIdentifyPlatformWorkflow(kc)
		require.NoError(t, err)
		require.NotNil(t, identifyPlatformWorkflow)

		m, err := NewManager(
			OptManagerLogger(logr.Discard()),
			OptManagerPeriod(time.Millisecond),
		)
		require.NoError(t, err)

		m.AddWorkflow(clusterStateWorkflow)
		m.AddWorkflow(identifyPlatformWorkflow)

		ch := make(chan Report)
		require.NoError(t, m.AddConsumer(ch))
		require.NoError(t, m.Start())

		report := <-ch
		m.Stop()

		require.EqualValues(t, Report{
			"cluster-state": provider.Report{
				"k8s-pod-count":     1,
				"k8s-service-count": 2,
			},
			"identify-platform": provider.Report{
				"k8s-cluster-arch":    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				"k8s-cluster-version": "v0.0.0-master+$Format:%H$",
				"k8s-provider":        provider.ClusterProviderUnknown,
			},
		}, report)
	})
}
