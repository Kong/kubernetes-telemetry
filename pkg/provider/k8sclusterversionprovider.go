package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
)

const (
	ClusterVersionKey  = "k8s-cluster-version"
	ClusterVersionKind = Kind(ClusterArchKey)
)

// NewK8sClusterVersionProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get cluster k8s version.
func NewK8sClusterVersionProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterVersionKind, kc, clusterVersionReport)
}

func clusterVersionReport(ctx context.Context, kc kubernetes.Interface) (Report, error) {
	v, err := clusterVersion(ctx, kc.Discovery())
	if err != nil {
		return nil, err
	}

	return Report{
		ClusterVersionKey: v,
	}, nil
}

// clusterVersion returns cluster's k8s version.
//
// NOTE:
// As of now it uses a simplified logic to GET the /version endpoint which
// might be OK for most use cases but for some, more granular approach might
// be needed to account for different versions of k8s nodes across the cluster.
func clusterVersion(ctx context.Context, d discovery.DiscoveryInterface) (string, error) {
	version, err := d.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster version: %w", err)
	}

	return (version.GitVersion), nil
}
