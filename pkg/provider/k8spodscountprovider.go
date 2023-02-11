package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// PodCountKey is report key under which the number of pods in the cluster
	// will be provided.
	PodCountKey = types.ProviderReportKey("k8s_pods_count")
	// PodCountKind represents the pod count provider kind.
	PodCountKind = Kind(PodCountKey)
)

// NewK8sPodCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a pod count from
// the cluster.
func NewK8sPodCountProvider(name string, d dynamic.Interface) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}

	return NewK8sObjectCountProvider(name, PodCountKind, d, gvr)
}
