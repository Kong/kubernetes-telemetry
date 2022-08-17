package telemetry

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/bombsimon/logrusr/v3"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
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
			&gatewayv1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "gateway-1",
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

		dynClient := dyn_fake.NewSimpleDynamicClientWithCustomListKinds(
			scheme.Scheme,
			map[schema.GroupVersionResource]string{
				{
					Group:    "gateway.networking.k8s.io",
					Version:  "v1beta1",
					Resource: "gateways",
				}: "GatewayList",
			},
			objs...,
		)

		clusterStateWorkflow, err := NewClusterStateWorkflow(dynClient, cl.RESTMapper())
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
				"k8s_nodes_count":    1,
				"k8s_pods_count":     1,
				"k8s_services_count": 2,
				// TODO: Even though we added Gateway API's schema to schema.Scheme, gateway count provider
				// doesn't detect it properly due to:
				// https://github.com/kubernetes/kubernetes/pull/110053.
				// When that's addressed we should revisit this test.
				// "k8s_gateways_count": 0,
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

func TestManagerWithMultilpleWorkflowsOneReturningError(t *testing.T) {
	logrusLog := logrus.New()
	logrusLog.Level = logrus.DebugLevel
	log := logrusr.New(logrusLog)
	m, err := NewManager(
		OptManagerLogger(log),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic")
		{
			p, err := provider.NewFixedValueProvider("constant1", provider.Report{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}
	{
		w := NewWorkflow("basic_with_error")
		{
			p, err := provider.NewFunctorProvider("error_provider",
				func(context.Context) (provider.Report, error) {
					return nil, errors.New("I am an error")
				},
			)
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

	ch := make(chan types.Report)
	consumer := NewRawConsumer(forwarders.NewRawChannelForwarder(ch))
	require.NoError(t, m.AddConsumer(consumer))
	require.NoError(t, m.Start())

	report := <-ch
	m.Stop()
	require.EqualValues(t, types.Report{
		"basic": provider.Report{
			"constant1": "value1",
		},
		"basic_with_error": provider.Report{
			"constant2": "value2",
		},
	}, report)
}
