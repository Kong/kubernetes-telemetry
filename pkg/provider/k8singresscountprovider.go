package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// IngressCountKey is report key under which the number of ingresses in the cluster
	// will be provided.
	IngressCountKey = types.ProviderReportKey("k8s_ingresses_count")
	// IngressCountKind represents the ingress count provider kind.
	IngressCountKind = Kind(IngressCountKey)
)

// NewK8sIngressCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a pod count from
// the cluster.
func NewK8sIngressCountProvider(name string, d dynamic.Interface) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "networking.k8s.io",
		Version:  "v1",
		Resource: "ingresses",
	}

	return NewK8sObjectCountProvider(name, IngressCountKind, d, gvr)
}
