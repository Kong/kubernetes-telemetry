package provider

import (
	"context"
	"strings"

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
		if p, ok := getClusterProviderFromVersion(cVersion.String()); ok {
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
		if p, ok := getClusterProviderFromNodes(nodeList); ok {
			return Report{
				ClusterProviderKey: p,
			}, nil
		}
	}

	return Report{
		ClusterProviderKey: ClusterProviderUnknown,
	}, nil
}

// getClusterProviderFromVersion tries to deduce the cluster provider based on
// the version string as returned by the /version API.
func getClusterProviderFromVersion(version string) (ClusterProvider, bool) {
	const (
		versionSubstringGKE = "gke"
		versionSubstringEKS = "eks"
	)

	if strings.Contains(version, versionSubstringGKE) {
		return ClusterProviderGKE, true
	}
	if strings.Contains(version, versionSubstringEKS) {
		return ClusterProviderAWS, true
	}

	return ClusterProviderUnknown, false
}

// clusterProviderDistribution represents a distribution of clusterproviders in
// a form of a map of cluster providers to the number of entities indicating
// a particular provider, e.g. nodes.
type clusterProviderDistribution map[ClusterProvider]int

func getClusterProviderFromNodes(nodeList *corev1.NodeList) (ClusterProvider, bool) {
	// Try to deduce cloud provider by looking at node provider ID field.
	if p, ok := getClusterProviderFromNodesProviderID(nodeList); ok {
		return p, true
	}

	// We still have not figured out which provider it is so try finding provider
	// specific labels and/or annotations on nodes.
	for _, n := range nodeList.Items {
		if p, ok := getClusterProviderFromAnnotations(n.Annotations); ok {
			return p, true
		}
		if p, ok := getClusterProviderFromLabels(n.Labels); ok {
			return p, true
		}
	}

	return ClusterProviderUnknown, false
}

func getClusterProviderFromNodesProviderID(nodeList *corev1.NodeList) (ClusterProvider, bool) {
	const (
		// Nodes on GKE are provided by GCE (Google Compute Engine)
		providerIDPrefixGKE = "gce"
		providerIDPrefixAWS = "aws"
	)
	d := make(clusterProviderDistribution)
	for _, n := range nodeList.Items {
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixGKE) {
			d[ClusterProviderGKE]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixAWS) {
			d[ClusterProviderAWS]++
			continue
		}
	}
	if p, ok := getMostCommonClusterProviderFromDistribution(d); ok {
		return p, true
	}
	return ClusterProviderUnknown, false
}

// getMostCommonClusterProviderFromDistribution returns the most commonly occurring
// cluster provider from the provided cluster provider distribution.
func getMostCommonClusterProviderFromDistribution(d clusterProviderDistribution) (ClusterProvider, bool) {
	var (
		top   = ClusterProviderUnknown
		found = false
		max   = 0
	)
	for k, v := range d {
		if v > max {
			top = k
			found = true
			max = v
		}
	}
	return top, found
}

func getClusterProviderFromAnnotations(annotations map[string]string) (ClusterProvider, bool) {
	const (
		annotationNameGKEInstanceID = "container.googleapis.com/instance_id"
	)

	annotationsGKE := map[string]struct{}{
		annotationNameGKEInstanceID: {},
	}

	// This approach currently loops through the provided annotations and checks
	// each of them against known annotations sets for particular cluster providers.
	// The reason for this is that regardless of the number of cluster providers
	// this will loop only once through the provided annotations (times O(1) map
	// lookup for each known annotations).
	for aName := range annotations {
		if _, ok := annotationsGKE[aName]; ok {
			return ClusterProviderGKE, true
		}
	}

	return ClusterProviderUnknown, false
}

func getClusterProviderFromLabels(labels map[string]string) (ClusterProvider, bool) {
	const (
		labelNameAWSClusterName = "alpha.eksctl.io/cluster-name"
		labelNameAWSInstanceID  = "alpha.eksctl.io/instance-id"
	)

	labelsAWS := map[string]struct{}{
		labelNameAWSClusterName: {},
		labelNameAWSInstanceID:  {},
	}

	// This approach currently loops through the provided labels and checks
	// each of them against known labels sets for particular cluster providers.
	// The reason for this is that regardless of the number of cluster providers
	// this will loop only once through the provided labels (times O(1) map
	// lookup for each known labels).
	for lName := range labels {
		if _, ok := labelsAWS[lName]; ok {
			return ClusterProviderAWS, true
		}
	}

	return ClusterProviderUnknown, false
}