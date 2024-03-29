package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// ServiceCountKey is report key under which the number of services in the cluster
	// will be provided.
	ServiceCountKey = types.ProviderReportKey("k8s_services_count")
	// ServiceCountKind represents the service count provider kind.
	ServiceCountKind = Kind(ServiceCountKey)
)

// NewK8sServiceCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a service count from
// the cluster.
func NewK8sServiceCountProvider(name string, d dynamic.Interface) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "services",
	}

	return NewK8sObjectCountProvider(name, ServiceCountKind, d, gvr)
}
