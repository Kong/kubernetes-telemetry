package provider

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/metadata"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// ReferenceGrantCountKey is report key under which the number of ReferenceGrants in the cluster
	// will be provided.
	ReferenceGrantCountKey = types.ProviderReportKey("k8s_referencegrants_count")
	// ReferenceGrantCountKind represents the ReferenceGrant count provider kind.
	ReferenceGrantCountKind = Kind(ReferenceGrantCountKey)
)

// NewK8sReferenceGrantCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a ReferenceGrant count from
// the cluster.
func NewK8sReferenceGrantCountProvider(name string, m metadata.Interface, rm meta.RESTMapper) (Provider, error) {
	gvr := schema.GroupVersionResource{
		Group:    "gateway.networking.k8s.io",
		Version:  "v1beta1",
		Resource: "referencegrants",
	}
	return NewK8sObjectCountProviderWithRESTMapper(name, ReferenceGrantCountKind, m, gvr, rm)
}
