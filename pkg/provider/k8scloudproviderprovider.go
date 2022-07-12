package provider

import (
	"context"
	"strings"

	"github.com/samber/mo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// ClusterProviderKey is report key under which the cluster provider will be provided.
	ClusterProviderKey = "k8s-provider"
	// ClusterProviderKind represents cluster provider kind.
	ClusterProviderKind = Kind(ClusterProviderKey)
)

// ClusterProvider identifies a particular clsuter provider like AWS, GKE, Azure etc.
type ClusterProvider string

const (
	// ClusterProviderGKE identifies Google's GKE cluster provider.
	ClusterProviderGKE = ClusterProvider("GKE")
	// ClusterProviderAWS identifies Amazon's AWS cluster provider.
	ClusterProviderAWS = ClusterProvider("AWS")
	// ClusterProviderUnknown represents an unknown cluster provider.
	ClusterProviderUnknown = ClusterProvider("UNKNOWN")
)

// NewK8sClusterProviderProvider creates telemetry data provider that will
// return the cluster provider name based on a set of heuristics.
func NewK8sClusterProviderProvider(name string, kc kubernetes.Interface) (Provider, error) {
	// TODO
	// return nil, fmt.Errorf("not implemented")
	return NewK8sClientGoBase(name, ClusterProviderKind, kc, clusterProviderReport)
}

func clusterProviderReport(ctx context.Context, kc kubernetes.Interface) (Report, error) {
	{
		// Try to figure out the cluster provider based on the version string
		// returned by the /version endpoint.

		cVersion, err := clusterVersion(ctx, kc.Discovery())
		if err != nil {
			return nil, err
		}
		if p, ok := getClusterProviderFromVersion(cVersion).Get(); ok {
			return Report{
				ClusterProviderKey: p,
			}, nil
		}
	}

	{
		// Try to figure out the cluster provider based on the spec, labels
		// and annotations set on cluster nodes.

		nodeList, err := kc.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		if p, ok := getClusterProviderFromNodes(nodeList).Get(); ok {
			return Report{
				ClusterProviderKey: p,
			}, nil
		}
	}

	return Report{
		ClusterProviderKey: ClusterProviderUnknown,
	}, nil
}

func getClusterProviderFromVersion(version string) mo.Option[ClusterProvider] {
	const (
		versionSubstringGKE = "gke"
		versionSubstringEKS = "eks"
	)

	if strings.Contains(version, versionSubstringGKE) {
		return mo.Some(ClusterProviderGKE)
	}
	if strings.Contains(version, versionSubstringEKS) {
		return mo.Some(ClusterProviderAWS)
	}

	return mo.None[ClusterProvider]()
}

func getClusterProviderFromNodes(nodeList *corev1.NodeList) mo.Option[ClusterProvider] {
	const (
		// Nodes on GKE are provided by GCE (Google Compute Engine)
		providerIDPrefixGKE = "gce"
		providerIDPrefixAWS = "aws"
	)

	m := make(map[ClusterProvider]int)
	for _, n := range nodeList.Items {
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixGKE) {
			m[ClusterProviderGKE]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixAWS) {
			m[ClusterProviderAWS]++
			continue
		}
	}

	// Just take the cluster provider that occurs the most often from the bunch
	// and return that.
	var (
		top = mo.None[ClusterProvider]()
		max = -1
	)
	for k, v := range m {
		if v > max {
			top = mo.Some(k)
			max = v
		}
	}
	if top.IsPresent() {
		return top
	}

	// We still have not figured out which provider it is so try finding provider
	// specific labels and/or annotations on nodes.

	for _, n := range nodeList.Items {
		if p, ok := getClusterProviderFromAnnotations(n.Annotations).Get(); ok {
			return mo.Some(p)
		}
		if p, ok := getClusterProviderFromLabels(n.Labels).Get(); ok {
			return mo.Some(p)
		}
	}

	return mo.None[ClusterProvider]()
}

func getClusterProviderFromAnnotations(annotations map[string]string) mo.Option[ClusterProvider] {
	const (
		annotationNameGKEInstanceID = "container.googleapis.com/instance_id"
	)

	annotationsGKE := map[string]struct{}{
		annotationNameGKEInstanceID: {},
	}

	for aName := range annotations {
		if _, ok := annotationsGKE[aName]; ok {
			return mo.Some(ClusterProviderGKE)
		}
	}

	return mo.None[ClusterProvider]()
}

func getClusterProviderFromLabels(labels map[string]string) mo.Option[ClusterProvider] {
	const (
		labelNameAWSClusterName = "alpha.eksctl.io/cluster-name"
		labelNameAWSInstanceID  = "alpha.eksctl.io/instance-id"
	)

	labelsAWS := map[string]struct{}{
		labelNameAWSClusterName: {},
		labelNameAWSInstanceID:  {},
	}

	for lName := range labels {
		if _, ok := labelsAWS[lName]; ok {
			return mo.Some(ClusterProviderAWS)
		}
	}

	return mo.None[ClusterProvider]()
}
