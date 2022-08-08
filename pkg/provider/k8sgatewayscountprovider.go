package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	// GatewayCountKey is report key under which the number of pods in the cluster
	// will be provided.
	GatewayCountKey = ReportKey("k8s_gateways_count")
	// GatewayCountKind represents the pod count provider kind.
	GatewayCountKind = Kind(GatewayCountKey)
)

// NewK8sGatewayCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a gateway count from
// the cluster.
func NewK8sGatewayCountProvider(name string, d dynamic.Interface) (Provider, error) {
	gvk := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1beta1",
		Resource: "gateways",
	}

	// TODO:
	// consider detecting what resource version is available on the cluster to
	// properly report. Alternatively consider reporting version together with
	// the count.
	return &k8sObjectCount{
		resource: d.Resource(gvk),
		gvk:      gvk,
		base: base{
			name: name,
			kind: GatewayCountKind,
		},
	}, nil
}
