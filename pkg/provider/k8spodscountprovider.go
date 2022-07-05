package provider

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodCountKey  = "k8s-pod-count"
	PodCountKind = Kind("k8s-pod-count")
)

// NewK8sPodCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get pod count from
// the cluster
func NewK8sPodCountProvider(name string, cl client.Client) (Provider, error) {
	return NewK8sControllerRuntimeBase(name, PodCountKind, cl, podCountReport)
}

func podCountReport(ctx context.Context, cl client.Client) (Report, error) {
	podsCount, err := podCount(ctx, cl)
	if err != nil {
		return nil, err
	}

	return Report{
		PodCountKey: podsCount,
	}, nil
}

func podCount(ctx context.Context, cl client.Client) (int, error) {
	var (
		podsList corev1.PodList
		count    int
	)

	for continueToken := ""; ; continueToken = podsList.Continue {
		err := cl.List(ctx, &podsList, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to list pods: %w", err)
		}

		count += len(podsList.Items)
		if podsList.Continue == "" {
			break
		}
	}

	return count, nil
}
