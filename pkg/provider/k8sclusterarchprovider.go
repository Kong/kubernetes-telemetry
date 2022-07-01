package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

type k8sClusterArch struct {
	// kc provides client-go's client implementation. This is used for retrieving
	// cluster's version and architecture through discovery interface.
	kc kubernetes.Interface

	base
}

var _ Provider = (*k8sClusterArch)(nil)

func NewK8sClusterArchProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return k8sClusterArch{
		kc: kc,
		base: base{
			name: name,
			kind: "k8s-cluster-arch",
		},
	}, nil
}

const (
	ClusterArchKey = "k8s-cluster-arch"
)

func (p k8sClusterArch) Provide(ctx context.Context) (Report, error) {
	cArch, err := p.clusterArch(ctx)
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
func (p k8sClusterArch) clusterArch(ctx context.Context) (string, error) {
	version, err := p.kc.Discovery().ServerVersion()
	if err != nil {
		return "", p.WrapError(fmt.Errorf("failed to get cluster architecture: %w", err))
	}

	return (version.Platform), nil
}
