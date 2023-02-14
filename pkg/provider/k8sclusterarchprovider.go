package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// ClusterArchKey is report key under which cluster architecture will be provided.
	ClusterArchKey = types.ProviderReportKey("k8s_arch")
	// ClusterArchKind represents cluster arch provider kind.
	ClusterArchKind = Kind(ClusterArchKey)
)

// NewK8sClusterArchProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get cluster architecture.
func NewK8sClusterArchProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterArchKind, kc, clusterArchReport)
}

func clusterArchReport(ctx context.Context, kc kubernetes.Interface) (types.ProviderReport, error) {
	cArch, err := clusterArch(ctx, kc.Discovery())
	if err != nil {
		return nil, err
	}

	return types.ProviderReport{
		ClusterArchKey: cArch,
	}, nil
}

// clusterArch returns cluster's architecture.
//
// NOTE:
// As of now it uses a simplified logic to GET the /version endpoint which
// might be OK for most use cases but for some, more granular approach might
// be needed to account for different versions/architectures of k8s nodes across
// the cluster.
func clusterArch(ctx context.Context, d discovery.DiscoveryInterface) (string, error) { //nolint:unparam
	version, err := d.ServerVersion()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster architecture: %w", err)
	}

	return (version.Platform), nil
}
