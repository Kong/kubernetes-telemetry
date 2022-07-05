package provider

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ServiceCountKey  = "k8s-service-count"
	ServiceCountKind = Kind("k8s-service-count")
)

// NewK8sServiceCountProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get service count from
// the cluster
func NewK8sServiceCountProvider(name string, cl client.Client) (Provider, error) {
	return NewK8sControllerRuntimeBase(name, ServiceCountKind, cl, serviceCountReport)
}

func serviceCountReport(ctx context.Context, cl client.Client) (Report, error) {
	servicesCount, err := serviceCount(ctx, cl)
	if err != nil {
		return nil, err
	}

	return Report{
		ServiceCountKey: servicesCount,
	}, nil
}

func serviceCount(ctx context.Context, cl client.Client) (int, error) {
	var (
		serviceList corev1.ServiceList
		count       int
	)

	for continueToken := ""; ; continueToken = serviceList.Continue {
		err := cl.List(ctx, &serviceList, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to list services: %w", err)
		}

		count += len(serviceList.Items)
		if serviceList.Continue == "" {
			break
		}
	}

	return count, nil
}
