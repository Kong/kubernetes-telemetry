package provider

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

type k8sClusterVersion struct {
	// kc provides client-go's client implementation. This is used for retrieving
	// cluster's version and architecture through discovery interface.
	kc kubernetes.Interface

	base
}

var _ Provider = (*k8sClusterVersion)(nil)

func NewK8sClusterVersionProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return k8sClusterVersion{
		kc: kc,
		base: base{
			name: name,
			kind: "k8s-cluster-version",
		},
	}, nil
}

const (
	ClusterVersionKey = "k8s-cluster-version"
)

func (p k8sClusterVersion) Provide(ctx context.Context) (Report, error) {
	cVersion, err := p.clusterVersion(ctx)
	if err != nil {
		return nil, err
	}

	return Report{
		ClusterVersionKey: cVersion,
	}, nil
}

// clusterVersion returns cluster's k8s version.
func (p k8sClusterVersion) clusterVersion(ctx context.Context) (string, error) {
	version, err := p.kc.Discovery().ServerVersion()
	if err != nil {
		return "", p.WrapError(fmt.Errorf("failed to get cluster version: %w", err))
	}

	return version.String(), nil
}
