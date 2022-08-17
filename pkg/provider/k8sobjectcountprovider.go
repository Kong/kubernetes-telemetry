package provider

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// k8sObjectCount is a Provider that returns the count of all objects of a certain kind in the Kubernetes cluster.
// Caller indicates object kind by passing its GroupVersionResource to `objectType`.
//
// Example: Use {Group: "", Version: "v1", Resource: "pods"} to get a Provider that counts all Pods in the cluster.
type k8sObjectCount struct {
	gvk      schema.GroupVersionResource
	resource dynamic.NamespaceableResourceInterface

	base
}

const (
	defaultPageLimit = 1000
)

func (k *k8sObjectCount) Provide(ctx context.Context) (Report, error) {
	var (
		count         int
		continueToken string
	)

	for {
		list, err := k.resource.List(ctx, metav1.ListOptions{
			// Conservatively use a limit for paging.
			Limit:    defaultPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return Report{}, k.WrapError(err)
		}

		count += len(list.Items)
		if continueToken = list.GetContinue(); continueToken == "" {
			break
		}
	}

	return Report{
		ReportKey("k8s_" + k.gvk.Resource + "_count"): count,
	}, nil
}
