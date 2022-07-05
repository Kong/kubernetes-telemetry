package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

const (
	ClusterArchKey  = "k8s-cluster-arch"
	ClusterArchKind = Kind("k8s-cluster-arch")
)

// NewK8sClusterArchProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get cluster architecture.
func NewK8sClusterArchProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterArchKind, kc, clusterArchReport)
}

func clusterArchReport(ctx context.Context, kc kubernetes.Interface) (Report, error) {
	cArch, err := clusterArch(ctx, kc)
	if err != nil {
		return nil, err
	}

	return Report{
		ClusterArchKey: cArch,
	}, nil
}

// clusterArch returns cluster's architecture.
// NOTE: As of now it uses a simplified logic to GET the /version endpoint which
//       might be OK for most use cases but for some, more granular approach might
//       be needed to account for different versions of k8s nodes across the cluster.
func clusterArch(ctx context.Context, kc kubernetes.Interface) (string, error) {
	version, err := kc.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster architecture: %w", err)
	}

	return (version.Platform), nil
}
