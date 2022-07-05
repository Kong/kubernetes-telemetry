package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

const (
	ClusterVersionKey  = "k8s-cluster-version"
	ClusterVersionKind = Kind("k8s-cluster-version")
)

// NewK8sClusterVersionProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get cluster k8s version.
func NewK8sClusterVersionProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterVersionKind, kc, clusterVersionReport)
}

func clusterVersionReport(ctx context.Context, kc kubernetes.Interface) (Report, error) {
	v, err := clusterVersion(ctx, kc)
	if err != nil {
		return nil, err
	}

	return Report{
		ClusterVersionKey: v,
	}, nil
}

// clusterVersion returns cluster's k8s version.
func clusterVersion(ctx context.Context, kc kubernetes.Interface) (string, error) {
	version, err := kc.Discovery().ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster version: %w", err)
	}

	return (version.Platform), nil
}
