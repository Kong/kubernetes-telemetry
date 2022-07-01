package provider

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sServicesCountProvider struct {
	// cl provides controller-runtime client implementation.
	cl client.Client

	base
}

var _ Provider = (*k8sServicesCountProvider)(nil)

func NewK8sServicesCountProvider(name string, cl client.Client) (Provider, error) {
	return k8sServicesCountProvider{
		cl: cl,
		base: base{
			name: name,
			kind: "k8s-service-count",
		},
	}, nil
}

const (
	ServiceCountKey = "k8s-service-count"
)

func (p k8sServicesCountProvider) Provide(ctx context.Context) (Report, error) {
	servicesCount, err := p.ServicesCount(ctx)
	if err != nil {
		return nil, err
	}

	return Report{
		ServiceCountKey: servicesCount,
	}, nil
}

// ServicesCount returns the number of services defined in the cluster.
func (p k8sServicesCountProvider) ServicesCount(ctx context.Context) (int, error) {
	var (
		serviceList corev1.ServiceList
		count       int
	)

	for continueToken := ""; ; continueToken = serviceList.Continue {
		err := p.cl.List(ctx, &serviceList, &client.ListOptions{
			Continue: continueToken,
		})
		if err != nil {
			return 0, p.WrapError(fmt.Errorf("failed to list services: %w", err))
		}

		count += len(serviceList.Items)
		if serviceList.Continue == "" {
			break
		}
	}

	return count, nil
}
