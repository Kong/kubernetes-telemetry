package provider

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// k8sObjectCount is a Provider that returns the count of all objects of a certain kind in the Kubernetes cluster.
// Caller indicates object kind by passing its GroupVersionResource to `objectType`.
//
// Example: Use {Group: "", Version: "v1", Resource: "pods"} to get a Provider that counts all Pods in the cluster.
type k8sObjectCount struct {
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
		list, err := k.resource.List(ctx, v1.ListOptions{
			// Conservatively use a limit for paging.
			Limit:    defaultPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return Report{}, err
		}

		count += len(list.Items)
		if continueToken = list.GetContinue(); continueToken == "" {
			break
		}
	}

	return Report{
		k.Name(): count,
	}, nil
}
