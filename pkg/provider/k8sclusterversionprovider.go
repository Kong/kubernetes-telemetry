package provider

import (
	"context"
	"fmt"

	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// ClusterVersionKey is the report key under which cluster k8s version will
	// be provided as returned by the /version API.
	ClusterVersionKey = types.ProviderReportKey("k8sv")
	// ClusterVersionSemverKey is the report key under which cluster k8s semver
	// version will be provided.
	ClusterVersionSemverKey = types.ProviderReportKey("k8sv_semver")
	// ClusterVersionKind represents cluster version provider kind.
	ClusterVersionKind = Kind(ClusterVersionKey)
)

// NewK8sClusterVersionProvider creates telemetry data provider that will query the
// configured k8s cluster - using the provided client - to get cluster k8s version.
func NewK8sClusterVersionProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterVersionKind, kc, clusterVersionReport)
}

func clusterVersionReport(ctx context.Context, kc kubernetes.Interface) (types.ProviderReport, error) {
	v, err := clusterVersion(ctx, kc.Discovery())
	if err != nil {
		return nil, err
	}

	semver, err := utilversion.ParseGeneric(v.GitVersion)
	if err != nil {
		// If we fail to decode the version then let's fall back to returning just
		// the major and minor returned from /version API.
		return types.ProviderReport{ //nolint:nilerr
			ClusterVersionKey:       v.GitVersion,
			ClusterVersionSemverKey: fmt.Sprintf("v%s.%s", v.Major, v.Minor),
		}, nil
	}

	return types.ProviderReport{
		ClusterVersionKey:       v.GitVersion,
		ClusterVersionSemverKey: "v" + semver.String(),
	}, nil
}

// clusterVersion returns cluster's k8s version.
//
// NOTE:
// As of now it uses a simplified logic to GET the /version endpoint which
// might be OK for most use cases but for some, more granular approach might
// be needed to account for different versions of k8s nodes across the cluster.
func clusterVersion(ctx context.Context, d discovery.DiscoveryInterface) (*version.Info, error) { //nolint:unparam
	v, err := d.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster version: %w", err)
	}

	return v, nil
}
