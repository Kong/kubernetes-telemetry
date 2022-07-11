package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodCountKey  = "k8s-pod-count"
	PodCountKind = Kind(PodCountKey)
)

// NewK8sPodCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a pod count from
// the cluster
func NewK8sPodCountProvider(name string, cl client.Client) Provider {
	return &k8sObjectCount{
		cl:         cl,
		objectType: schema.GroupVersionKind{Group: "", Kind: "PodList", Version: "v1"},
		base:       base{name: name, kind: PodCountKind},
	}
}
