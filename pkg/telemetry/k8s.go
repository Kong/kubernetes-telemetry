package telemetry

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/Kong/kubernetes-telemetry/pkg/provider"
)

const (
	// IdentifyPlatformWorkflowName is the name assigned to identify platform
	// workflow.
	IdentifyPlatformWorkflowName = "identify-platform"
)

// NewIdentifyPlatformWorkflow creates a new 'identify-platform' workflow, based
// on a predefined set of providers that will deliver telemetry data from a cluster.
//
// Exemplar report produced:
//
//	{
//	  "k8s-cluster-arch": "linux/amd64",
//	  "k8s-cluster-version": "v1.24.1-gke.1400",
//	  "k8s-cluster-version-semver": "v1.24.1",
//	  "k8s-provider": "GKE"
//	}
func NewIdentifyPlatformWorkflow(kc kubernetes.Interface) (Workflow, error) {
	if kc == nil {
		return nil, ErrNilKubernetesInterfaceProvided
	}

	pClusterArch, err := provider.NewK8sClusterArchProvider(provider.ClusterArchKey, kc)
	if err != nil {
		return nil, err
	}

	pClusterVersion, err := provider.NewK8sClusterVersionProvider(provider.ClusterVersionKey, kc)
	if err != nil {
		return nil, err
	}

	pClusterProvider, err := provider.NewK8sClusterProviderProvider(provider.ClusterProviderKey, kc)
	if err != nil {
		return nil, err
	}

	w := NewWorkflow(IdentifyPlatformWorkflowName)
	w.AddProvider(pClusterArch)
	w.AddProvider(pClusterVersion)
	w.AddProvider(pClusterProvider)

	return w, nil
}

const (
	// ClusterStateWorkflowName is the name assigned to cluster state workflow.
	ClusterStateWorkflowName = "cluster-state"
)

// NewClusterStateWorkflow creates a new 'cluster-state' workflow, based on a predefined
// set of providers that will deliver telemetry data about the cluster state.
//
// Exemplar report produced:
//
//	{
//	  "k8s-pod-count": 21,
//	  "k8s-service-count": 3
//	}
func NewClusterStateWorkflow(d dynamic.Interface) (Workflow, error) {
	if d == nil {
		return nil, ErrNilDynClientProvided
	}

	providerPodCount, err := provider.NewK8sPodCountProvider(provider.PodCountKey, d)
	if err != nil {
		return nil, err
	}
	providerServiceCount, err := provider.NewK8sServiceCountProvider(provider.ServiceCountKey, d)
	if err != nil {
		return nil, err
	}

	w := NewWorkflow(ClusterStateWorkflowName)
	w.AddProvider(providerPodCount)
	w.AddProvider(providerServiceCount)

	return w, nil
}
