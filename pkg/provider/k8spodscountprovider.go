package provider

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sPodsCountProvider struct {
	// cl provides controller-runtime client implementation.
	cl client.Client

	base
}

var _ Provider = (*k8sPodsCountProvider)(nil)

func NewK8sPodsCountProvider(name string, cl client.Client) (Provider, error) {
	return k8sPodsCountProvider{
		cl: cl,
		base: base{
			name: name,
			kind: "k8s-pod-count",
		},
	}, nil
}

const (
	PodCountKey = "k8s-pod-count"
)

func (p k8sPodsCountProvider) Provide(ctx context.Context) (Report, error) {
	podsCount, err := p.PodsCount(ctx)
	if err != nil {
		return nil, err
	}

	return Report{
		PodCountKey: podsCount,
	}, nil
}

// PodsCount returns the number of pods defined in the cluster.
func (p k8sPodsCountProvider) PodsCount(ctx context.Context) (int, error) {
	var (
		podsList corev1.PodList
		count    int
	)

	for continueToken := ""; ; continueToken = podsList.Continue {
		err := p.cl.List(ctx, &podsList, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return 0, p.WrapError(fmt.Errorf("failed to list pods: %w", err))
		}

		count += len(podsList.Items)
		if podsList.Continue == "" {
			break
		}
	}

	return count, nil
}
