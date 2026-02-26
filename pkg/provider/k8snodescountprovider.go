package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// NodeCountKey is the report key under which the number of nodes in the cluster
	// will be provided.
	NodeCountKey = types.ProviderReportKey("k8s_nodes_count")
	// NodeCountKind represents the node count provider kind.
	NodeCountKind = Kind(NodeCountKey)
)

// NewK8sNodeCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a node count from
// the cluster.
func NewK8sNodeCountProvider(name string, m metadata.Interface) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}

	return NewK8sObjectCountProvider(name, NodeCountKind, m, gvr)
}
