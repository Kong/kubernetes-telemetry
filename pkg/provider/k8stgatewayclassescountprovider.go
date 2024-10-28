package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// GatewayClassCountKey is report key under which the number of GatewayClasses in the cluster
	// will be provided.
	GatewayClassCountKey = types.ProviderReportKey("k8s_gatewayclasses_count")
	// GatewayClassCountKind represents the GatewayClass count provider kind.
	GatewayClassCountKind = Kind(GatewayClassCountKey)
)

// NewK8sGatewayClassCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a GatewayClass count from
// the cluster.
func NewK8sGatewayClassCountProvider(name string, d dynamic.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1",
		Resource: "gatewayclasses",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, GatewayClassCountKind, d, gvr, rm)
}
