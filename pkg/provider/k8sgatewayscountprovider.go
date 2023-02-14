package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// GatewayCountKey is report key under which the number of pods in the cluster
	// will be provided.
	GatewayCountKey = types.ProviderReportKey("k8s_gateways_count")
	// GatewayCountKind represents the pod count provider kind.
	GatewayCountKind = Kind(GatewayCountKey)
)

// NewK8sGatewayCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a gateway count from
// the cluster.
func NewK8sGatewayCountProvider(name string, d dynamic.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1beta1",
		Resource: "gateways",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, GatewayCountKind, d, gvr, rm)
}
