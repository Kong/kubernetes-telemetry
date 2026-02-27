package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// UDPRouteCountKey is report key under which the number of UDPRoutes in the cluster
	// will be provided.
	UDPRouteCountKey = types.ProviderReportKey("k8s_udproutes_count")
	// UDPRouteCountKind represents the UDPRoute count provider kind.
	UDPRouteCountKind = Kind(UDPRouteCountKey)
)

// NewK8sUDPRouteCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a UDPRoute count from
// the cluster.
func NewK8sUDPRouteCountProvider(name string, m metadata.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1alpha2",
		Resource: "udproutes",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, UDPRouteCountKind, m, gvr, rm)
}
