package provider

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// k8sObjectCount is a Provider that returns the count of all objects of a certain kind in the Kubernetes cluster.
// Caller indicates object kind by passing its "*List" GVK to `objectType`.
//
// Example: Use {Group: "", Version: "v1", Kind: "PodList"} to get a Provider that counts all Pods in the cluster.
type k8sObjectCount struct {
	cl         client.Client
	objectType schema.GroupVersionKind

	base
}

func (k *k8sObjectCount) Provide(ctx context.Context) (Report, error) {
	var (
		list        unstructured.UnstructuredList
		resultCount int
	)

	list.SetGroupVersionKind(k.objectType)

	// We could consider using the ListMeta field RemainingItemCount instead of iterating, but per v1.24 documentation
	// it's not guaranteed to be accurate.
	for continueToken := ""; ; continueToken = list.GetContinue() {
		err := k.cl.List(ctx, &list, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return nil, k.WrapError(fmt.Errorf("failed to list %v: %w", k.objectType.String(), err))
		}

		resultCount += len(list.Items)
		if list.GetContinue() == "" {
			break
		}
	}

	return Report{
		k.Name(): resultCount,
	}, nil
}
