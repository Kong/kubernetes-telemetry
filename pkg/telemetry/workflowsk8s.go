package telemetry

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
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
//	  "k8sv": "linux/amd64",
//	  "k8sv": "v1.24.1-gke.1400",
//	  "k8sv_semver": "v1.24.1",
//	  "k8s_provider": "GKE"
//	}
func NewIdentifyPlatformWorkflow(kc kubernetes.Interface) (Workflow, error) {
	if kc == nil {
		return nil, ErrNilKubernetesInterfaceProvided
	}

	pClusterArch, err := provider.NewK8sClusterArchProvider(string(provider.ClusterArchKey), kc)
	if err != nil {
		return nil, err
	}

	pClusterVersion, err := provider.NewK8sClusterVersionProvider(string(provider.ClusterVersionKey), kc)
	if err != nil {
		return nil, err
	}

	pClusterProvider, err := provider.NewK8sClusterProviderProvider(string(provider.ClusterProviderKey), kc)
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
// When a non-builtin CRD (like Gateway from Gateway API) is not available then
// the provider for this resource's telemetry data is not added to the workflow.
//
// Exemplar report produced:
//
//	{
//	  "k8s_pods_count": 21,
//	  "k8s_services_count": 3,
//	  "k8s_gateways_count": 1,
//	  "k8s_nodes_count": 1
//	}
func NewClusterStateWorkflow(d dynamic.Interface, rm meta.RESTMapper) (Workflow, error) {
	if d == nil {
		return nil, ErrNilDynClientProvided
	}

	w := NewWorkflow(ClusterStateWorkflowName)

	providerPodCount, err := provider.NewK8sPodCountProvider(string(provider.PodCountKey), d)
	if err != nil {
		return nil, err
	}
	w.AddProvider(providerPodCount)

	providerServiceCount, err := provider.NewK8sServiceCountProvider(string(provider.ServiceCountKey), d)
	if err != nil {
		return nil, err
	}
	w.AddProvider(providerServiceCount)

	providerNodeCount, err := provider.NewK8sNodeCountProvider(string(provider.NodeCountKey), d)
	if err != nil {
		return nil, err
	}
	w.AddProvider(providerNodeCount)

	// Below listed count providers are added optionally, when the corresponding CRDs are present
	// in the cluster.
	providerGatewayCount, err := provider.NewK8sGatewayCountProvider(string(provider.GatewayCountKey), d, rm)
	if err != nil {
		if !meta.IsNoMatchError(err) {
			return nil, err
		}
		// If there's no kind for Gateway in the cluster then just don't add its count provider.
	} else {
		w.AddProvider(providerGatewayCount)
	}

	return w, nil
}
