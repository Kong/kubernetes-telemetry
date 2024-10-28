package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// GRPCRouteCountKey is report key under which the number of GRPCRoutes in the cluster
	// will be provided.
	GRPCRouteCountKey = types.ProviderReportKey("k8s_grpcroutes_count")
	// GRPCRouteCountKind represents the GRPCRoute count provider kind.
	GRPCRouteCountKind = Kind(GRPCRouteCountKey)
)

// NewK8sGRPCRouteCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a GRPCRoute count from
// the cluster.
func NewK8sGRPCRouteCountProvider(name string, d dynamic.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "grpcroutes",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, GRPCRouteCountKind, d, gvr, rm)
}
