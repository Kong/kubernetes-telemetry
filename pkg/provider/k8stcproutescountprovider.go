package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// TCPRouteCountKey is report key under which the number of TCPRoutes in the cluster
	// will be provided.
	TCPRouteCountKey = types.ProviderReportKey("k8s_tcproutes_count")
	// TCPRouteCountKind represents the TCPRoute count provider kind.
	TCPRouteCountKind = Kind(TCPRouteCountKey)
)

// NewK8sTCPRouteCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a TCPRoute count from
// the cluster.
func NewK8sTCPRouteCountProvider(name string, d dynamic.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1alpha2",
		Resource: "tcproutes",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, TCPRouteCountKind, d, gvr, rm)
}
