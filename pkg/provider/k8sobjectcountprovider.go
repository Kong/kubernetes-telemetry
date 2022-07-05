package provider

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sObjectCount struct {
	cl         client.Client
	objectType schema.GroupVersionKind

	base
}

func (k *k8sObjectCount) Provide(ctx context.Context) (Report, error) {
	count, err := objectCount(ctx, k.cl, k.objectType)
	if err != nil {
		return nil, k.WrapError(err)
	}

	return Report{
		k.Name(): count,
	}, nil
}

func objectCount(ctx context.Context, cl client.Client, gvk schema.GroupVersionKind) (int, error) {
	var (
		list  unstructured.UnstructuredList
		count int
	)

	list.SetGroupVersionKind(gvk)

	for continueToken := ""; ; continueToken = list.GetContinue() {
		err := cl.List(ctx, &list, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to list %v: %w", gvk.String(), err)
		}

		count += len(list.Items)
		if list.GetContinue() == "" {
			break
		}
	}

	return count, nil
}