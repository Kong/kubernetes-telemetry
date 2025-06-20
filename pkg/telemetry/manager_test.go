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
	"go.uber.org/goleak"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	dyn_fake "k8s.io/client-go/dynamic/fake"
	clientgo_fake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	ctrlclient_fake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/kong/kubernetes-telemetry/pkg/forwarders"
	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/serializers"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

func TestManagerStartStopDoesntFail(t *testing.T) {
	m, err := NewManager("dummy-signal", OptManagerLogger(logr.Discard()))
	require.NoError(t, err)
	require.NoError(t, m.Start())
	m.Stop()
}

func TestManagerBasicLogicWorks(t *testing.T) {
	m, err := NewManager(
		"dummy-signal",
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
		{
			p, err := provider.NewFixedValueProvider("constant1", types.ProviderReport{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", types.ProviderReport{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	consumer := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewDiscardForwarder())

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
	require.EqualValues(t,
		types.SignalReport{
			Report: types.Report{
				"basic1": types.ProviderReport{
					"constant1": "value1",
					"constant2": "value2",
				},
			},
			Signal: "dummy-signal",
		}, report,
	)
}

func TestManagerWithMultilpleWorkflows(t *testing.T) {
	m, err := NewManager(
		"dummy-signal",
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
		{
			p, err := provider.NewFixedValueProvider("constant1", types.ProviderReport{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", types.ProviderReport{
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
			p, err := provider.NewFixedValueProvider("constant1", types.ProviderReport{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", types.ProviderReport{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	consumer := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewDiscardForwarder())
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
	require.EqualValues(t,
		types.SignalReport{
			Report: types.Report{
				"basic1": types.ProviderReport{
					"constant1": "value1",
					"constant2": "value2",
				},
				"basic2": types.ProviderReport{
					"constant1": "value1",
					"constant2": "value2",
				},
			},
			Signal: "dummy-signal",
		},
		report,
	)
}

func TestManagerWithCatalogWorkflows(t *testing.T) {
	t.Run("identify platform and cluster state", func(t *testing.T) {
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
			&netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "ingress-1",
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

		cl := ctrlclient_fake.
			NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithRuntimeObjects(objs...).
			Build()

		dynClient := dyn_fake.NewSimpleDynamicClient(cl.Scheme(), objs...)

		clusterStateWorkflow, err := NewClusterStateWorkflow(dynClient, cl.RESTMapper())
		require.NoError(t, err)
		require.NotNil(t, clusterStateWorkflow)

		kc := clientgo_fake.NewSimpleClientset()
		identifyPlatformWorkflow, err := NewIdentifyPlatformWorkflow(kc)
		require.NoError(t, err)
		require.NotNil(t, identifyPlatformWorkflow)

		m, err := NewManager(
			"dummy-signal",
			OptManagerLogger(logr.Discard()),
			OptManagerPeriod(time.Millisecond),
		)
		require.NoError(t, err)

		m.AddWorkflow(clusterStateWorkflow)
		m.AddWorkflow(identifyPlatformWorkflow)

		consumer := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewDiscardForwarder())
		require.NoError(t, m.AddConsumer(consumer))
		require.NoError(t, m.Start())

		report := <-consumer.ch
		m.Stop()

		require.Equal(t,
			types.SignalReport{
				Report: types.Report{
					"cluster-state": types.ProviderReport{
						"k8s_nodes_count":     1,
						"k8s_pods_count":      1,
						"k8s_services_count":  2,
						"k8s_ingresses_count": 1,
						// TODO: Even though we added Gateway API's schema to schema.Scheme, the below count providers
						// don't detect they properly due to https://github.com/kubernetes/kubernetes/pull/110053.
						// When that's addressed we should revisit this test and adjust test to check for
						// provider.GatewayClassCountKey:   1,
						// provider.GatewayCountKey:        1,
						// provider.HTTPRouteCountKey:      1,
						// provider.ReferenceGrantCountKey: 1,
						// provider.GRPCRouteCountKey:      1,
						// provider.TCPRouteCountKey:       1,
						// provider.UDPRouteCountKey:       1,
						// provider.TLSRouteCountKey:       1,
					},
					"identify-platform": types.ProviderReport{
						"k8s_arch":     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
						"k8sv":         "v0.0.0-master+$Format:%H$",
						"k8sv_semver":  "v0.0.0",
						"k8s_provider": provider.ClusterProviderUnknown,
					},
				},
				Signal: "dummy-signal",
			},
			report,
		)
	})
}

func TestManagerTriggerExecute(t *testing.T) {
	t.Run("TriggerExecute successfully triggers an execution", func(t *testing.T) {
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
			&netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "kong",
					Name:      "ingress-1",
				},
			},
			&corev1.Node{
				Spec: corev1.NodeSpec{
					ProviderID: "aws:///eu-west-1b/i-0fa11111111111111",
				},
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
		dynClient := dyn_fake.NewSimpleDynamicClient(scheme.Scheme, objs...)

		clusterStateWorkflow, err := NewClusterStateWorkflow(dynClient, cl.RESTMapper())
		require.NoError(t, err)
		require.NotNil(t, clusterStateWorkflow)

		kc := clientgo_fake.NewSimpleClientset(objs...)
		identifyPlatformWorkflow, err := NewIdentifyPlatformWorkflow(kc)
		require.NoError(t, err)
		require.NotNil(t, identifyPlatformWorkflow)

		m, err := NewManager(
			"dummy-signal",
			OptManagerLogger(logr.Discard()),
			OptManagerPeriod(time.Hour),
		)
		require.NoError(t, err)

		m.AddWorkflow(clusterStateWorkflow)
		m.AddWorkflow(identifyPlatformWorkflow)

		ch := make(chan []byte)
		consumer := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewChannelForwarder(ch))
		require.NoError(t, m.AddConsumer(consumer))
		require.NoError(t, m.Start())
		require.NoError(t, m.TriggerExecute(context.Background(), "ping"))

		report := <-ch
		m.Stop()

		arch := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
		require.EqualValues(t,
			fmt.Sprintf("<14>signal=ping;k8s_arch=%s;k8s_provider=AWS;k8sv=v0.0.0-master+$Format:%%H$;k8sv_semver=v0.0.0;k8s_ingresses_count=1;k8s_nodes_count=1;k8s_pods_count=1;k8s_services_count=2;\n", arch),
			string(report),
		)
	})
}

func TestManagerWithMultilpleWorkflowsOneReturningError(t *testing.T) {
	logrusLog := logrus.New()
	logrusLog.Level = logrus.DebugLevel
	log := logrusr.New(logrusLog)
	m, err := NewManager(
		"dummy-signal",
		OptManagerLogger(log),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic")
		{
			p, err := provider.NewFixedValueProvider("constant1", types.ProviderReport{
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
				func(context.Context) (types.ProviderReport, error) {
					return nil, errors.New("I am an error")
				},
			)
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{

			p, err := provider.NewFixedValueProvider("constant2", types.ProviderReport{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	ch := make(chan types.SignalReport)
	consumer := NewRawConsumer(forwarders.NewRawChannelForwarder(ch))
	require.NoError(t, m.AddConsumer(consumer))
	require.NoError(t, m.Start())

	report := <-ch
	m.Stop()
	require.EqualValues(t,
		types.SignalReport{
			Report: types.Report{
				"basic": types.ProviderReport{
					"constant1": "value1",
				},
				"basic_with_error": types.ProviderReport{
					"constant2": "value2",
				},
			},
			Signal: "dummy-signal",
		},
		report,
	)
}

func TestManagerNoGoroutineLeak(t *testing.T) {
	t.Cleanup(func() {
		t.Logf("Checking goroutine leak")
		// Verify that all goroutines started by the manager have been closed.
		goleak.VerifyNone(t)
	})

	m, err := NewManager(
		"dummy-signal",
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
		{
			p, err := provider.NewFixedValueProvider("constant1", types.ProviderReport{
				"constant1": "value1",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}
		{
			p, err := provider.NewFixedValueProvider("constant2", types.ProviderReport{
				"constant2": "value2",
			})
			require.NoError(t, err)
			w.AddProvider(p)
		}

		m.AddWorkflow(w)
	}

	consumer1 := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewDiscardForwarder())
	consumer2 := NewConsumer(serializers.NewSemicolonDelimited(), forwarders.NewDiscardForwarder())

	require.NoError(t, m.AddConsumer(consumer1))
	require.NoError(t, m.AddConsumer(consumer2))
	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)

	require.ErrorIs(t, m.AddConsumer(consumer1),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	// Stop manager.
	m.Stop()
}
