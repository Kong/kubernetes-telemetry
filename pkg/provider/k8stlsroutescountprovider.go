package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// TLSRouteCountKey is report key under which the number of TLSRoutes in the cluster
	// will be provided.
	TLSRouteCountKey = types.ProviderReportKey("k8s_tlsroutes_count")
	// TLSRouteCountKind represents the TLSRoute count provider kind.
	TLSRouteCountKind = Kind(TLSRouteCountKey)
)

// NewK8sTLSRouteCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a TLSRoute count from
// the cluster.
func NewK8sTLSRouteCountProvider(name string, m metadata.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1alpha2",
		Resource: "tlsroutes",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, TLSRouteCountKind, m, gvr, rm)
}
