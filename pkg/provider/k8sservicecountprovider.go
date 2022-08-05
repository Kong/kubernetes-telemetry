package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	// ServiceCountKey is report key under which the number of services in the cluster
	// will be provided.
	ServiceCountKey = ReportKey("k8s_services_count")
	// ServiceCountKind represents the service count provider kind.
	ServiceCountKind = Kind(ServiceCountKey)
)

// NewK8sServiceCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a service count from
// the cluster.
func NewK8sServiceCountProvider(name string, d dynamic.Interface) (Provider, error) {
	gvk := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	return &k8sObjectCount{
		resource: d.Resource(gvk),
		gvk:      gvk,
		base: base{
			name: name,
			kind: ServiceCountKind,
		},
	}, nil
}
