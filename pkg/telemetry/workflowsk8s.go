package telemetry

import (
	"errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
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
//	  "k8s_nodes_count": 1
//	  "k8s_gatewayclasses_count": 1,
//	  "k8s_gateways_count": 1,
//	  "k8s_httproutes_count": 1,
//	  "k8s_grpcroutes_count": 1,
//	  "k8s_tlsroutes_count": 1,
//	  "k8s_tcproutes_count": 1,
//	  "k8s_udproutes_count": 1,
//	  "k8s_referencegrants_count": 1
//	}
func NewClusterStateWorkflow(d dynamic.Interface, rm meta.RESTMapper) (Workflow, error) {
	if d == nil {
		return nil, ErrNilDynClientProvided
	}

	w := NewWorkflow(ClusterStateWorkflowName)

	coreObjects := []struct {
		providerCreator func(string, dynamic.Interface) (provider.Provider, error)
		countKey        types.ProviderReportKey
	}{
		{
			provider.NewK8sPodCountProvider,
			provider.PodCountKey,
		},
		{
			provider.NewK8sServiceCountProvider,
			provider.ServiceCountKey,
		},
		{
			provider.NewK8sNodeCountProvider,
			provider.NodeCountKey,
		},
	}
	for _, co := range coreObjects {
		provider, err := co.providerCreator(string(co.countKey), d)
		if err != nil {
			return nil, err
		}
		w.AddProvider(provider)
	}

	// Below listed count providers for resources from API group "gateway.networking.k8s.io",
	// are added optionally, when the corresponding CRDs are present in the cluster.
	optionalObjects := []struct {
		providerCreator func(string, dynamic.Interface, meta.RESTMapper) (provider.Provider, error)
		countKey        types.ProviderReportKey
	}{
		{
			provider.NewK8sGatewayClassCountProvider,
			provider.GatewayClassCountKey,
		},
		{
			provider.NewK8sGatewayCountProvider,
			provider.GatewayCountKey,
		},
		{
			provider.NewK8sHTTPRouteCountProvider,
			provider.HTTPRouteCountKey,
		},
		{
			provider.NewK8sGRPCRouteCountProvider,
			provider.GRPCRouteCountKey,
		},
		{
			provider.NewK8sTLSRouteCountProvider,
			provider.TLSRouteCountKey,
		},
		{
			provider.NewK8sTCPRouteCountProvider,
			provider.TCPRouteCountKey,
		},
		{
			provider.NewK8sUDPRouteCountProvider,
			provider.UDPRouteCountKey,
		},
		{
			provider.NewK8sReferenceGrantCountProvider,
			provider.ReferenceGrantCountKey,
		},
	}
	for _, oo := range optionalObjects {
		p, err := oo.providerCreator(string(oo.countKey), d, rm)
		if err != nil {
			if errGVR := (provider.ErrGVRNotAvailable{}); errors.As(err, &errGVR) {
				// GVR unavailable, just skip it.
				continue
			}
			return nil, err
		}
		w.AddProvider(p)
	}

	return w, nil
}
