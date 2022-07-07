package provider

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceCountKey  = "k8s-service-count"
	ServiceCountKind = Kind(ServiceCountKey)
)

// NewK8sServiceCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get a service count from
// the cluster
func NewK8sServiceCountProvider(name string, cl client.Client) Provider {
	return &k8sObjectCount{
		cl:         cl,
		objectType: schema.GroupVersionKind{Group: "", Kind: "ServiceList", Version: "v1"},
		base:       base{name: name, kind: ServiceCountKind},
	}
}
