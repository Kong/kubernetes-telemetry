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

	"github.com/kong/kubernetes-telemetry/pkg/forwarders"
	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/serializers"
	"github.com/kong/kubernetes-telemetry/pkg/types"
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

	consumer := NewConsumer(serializers.NewSemicolonDelimited("ping"), forwarders.NewDiscardForwarder())

	require.NoError(t, m.AddConsumer(consumer))
	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)
	require.ErrorIs(t, m.AddConsumer(consumer),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	report := <-consumer.ch
	m.Stop()
	require.EqualValues(t, types.Report{
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

	consumer := NewConsumer(serializers.NewSemicolonDelimited("ping"), forwarders.NewDiscardForwarder())
	require.NoError(t, m.AddConsumer(consumer))

	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)
	require.ErrorIs(t, m.AddConsumer(consumer),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	ch := consumer.ch
	report := <-ch
	m.Stop()
	require.EqualValues(t, types.Report{
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

		consumer := NewConsumer(serializers.NewSemicolonDelimited("ping"), forwarders.NewDiscardForwarder())
		require.NoError(t, m.AddConsumer(consumer))
		require.NoError(t, m.Start())

		report := <-consumer.ch
		m.Stop()

		require.EqualValues(t, types.Report{
			"cluster-state": provider.Report{
				"k8s_pods_count":     1,
				"k8s_services_count": 2,
			},
			"identify-platform": provider.Report{
				"k8s_arch":     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				"k8sv":         "v0.0.0-master+$Format:%H$",
				"k8sv_semver":  "v0.0.0",
				"k8s_provider": provider.ClusterProviderUnknown,
			},
		}, report)
	})
}
