package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// HTTPRouteCountKey is report key under which the number of HTTPRoutes in the cluster
	// will be provided.
	HTTPRouteCountKey = types.ProviderReportKey("k8s_httproutes_count")
	// HTTPRouteCountKind represents the HTTPRoute count provider kind.
	HTTPRouteCountKind = Kind(HTTPRouteCountKey)
)

// NewK8sHTTPRouteCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a HTTPRoute count from
// the cluster.
func NewK8sHTTPRouteCountProvider(name string, d dynamic.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "httproutes",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, HTTPRouteCountKind, d, gvr, rm)
}
