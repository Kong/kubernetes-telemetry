package provider

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// ClusterProviderKey is report key under which the cluster provider will be provided.
	ClusterProviderKey = types.ProviderReportKey("k8s_provider")
	// ClusterProviderKind represents cluster provider kind.
	ClusterProviderKind = Kind(ClusterProviderKey)
)

// ClusterProvider identifies a particular clsuter provider like AWS, GKE, Azure etc.
type ClusterProvider string

const (
	// ClusterProviderGKE identifies Google's GKE cluster provider.
	ClusterProviderGKE = ClusterProvider("GKE")
	// ClusterProviderAzure identifies Microsoft's Azure cluster provider.
	ClusterProviderAzure = ClusterProvider("Azure")
	// ClusterProviderAWS identifies Amazon's AWS cluster provider.
	ClusterProviderAWS = ClusterProvider("AWS")
	// ClusterProviderKubernetesInDocker identifies kind (kubernetes in docker) as cluster provider.
	ClusterProviderKubernetesInDocker = ClusterProvider("kind")
	// ClusterProviderK3S identifies k3s cluster provider.
	ClusterProviderK3S = ClusterProvider("k3s")
	// ClusterProviderRKE2 identifies RKE2 cluster provider.
	ClusterProviderRKE2 = ClusterProvider("rke2")
	// ClusterProviderUnknown represents an unknown cluster provider.
	ClusterProviderUnknown = ClusterProvider("UNKNOWN")
)

// NewK8sClusterProviderProvider creates telemetry data provider that will
// return the cluster provider name based on a set of heuristics.
func NewK8sClusterProviderProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, ClusterProviderKind, kc, clusterProviderReport)
}

func clusterProviderReport(ctx context.Context, kc kubernetes.Interface) (types.ProviderReport, error) {
	{
		// Try to figure out the cluster provider based on the version string
		// returned by the /version endpoint.

		cVersion, err := clusterVersion(ctx, kc.Discovery())
		if err != nil {
			return nil, err
		}
		if p, ok := getClusterProviderFromVersion(cVersion.String()); ok {
			return types.ProviderReport{
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
			return types.ProviderReport{
				ClusterProviderKey: p,
			}, nil
		}
	}

	return types.ProviderReport{
		ClusterProviderKey: ClusterProviderUnknown,
	}, nil
}

// getClusterProviderFromVersion tries to deduce the cluster provider based on
// the version string as returned by the /version API.
func getClusterProviderFromVersion(version string) (ClusterProvider, bool) {
	const (
		versionSubstringGKE  = "gke"
		versionSubstringEKS  = "eks"
		versionSubstringRKE2 = "rke2"
	)

	if strings.Contains(version, versionSubstringGKE) {
		return ClusterProviderGKE, true
	}
	if strings.Contains(version, versionSubstringEKS) {
		return ClusterProviderAWS, true
	}
	if strings.Contains(version, versionSubstringRKE2) {
		return ClusterProviderRKE2, true
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
		providerIDPrefixGKE   = "gce"
		providerIDPrefixAWS   = "aws"
		providerIDPrefixKind  = "kind"
		providerIDPrefixK3s   = "k3s"
		providerIDPrefixAzure = "azure"
	)

	d := make(clusterProviderDistribution)
	for _, n := range nodeList.Items {
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixGKE) {
			d[ClusterProviderGKE]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixAzure) {
			d[ClusterProviderAzure]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixAWS) {
			d[ClusterProviderAWS]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixKind) {
			d[ClusterProviderKubernetesInDocker]++
			continue
		}
		if strings.HasPrefix(n.Spec.ProviderID, providerIDPrefixK3s) {
			d[ClusterProviderK3S]++
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
	annotationsGKE := map[string]struct{}{
		"container.googleapis.com/instance_id": {},
	}

	annotationsK3S := map[string]struct{}{
		"k3s.io/hostname":         {},
		"k3s.io/internal-ip":      {},
		"k3s.io/node-args":        {},
		"k3s.io/node-config-hash": {},
		"k3s.io/node-env":         {},
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
		if _, ok := annotationsK3S[aName]; ok {
			return ClusterProviderK3S, true
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

	const (
		labelNameAzureClusterName      = "kubernetes.azure.com/cluster"
		labelNameAzureOSSKU            = "kubernetes.azure.com/os-sku"
		labelNameAzureRole             = "kubernetes.azure.com/role"
		labelNameAzureNodeImageVersion = "kubernetes.azure.com/node-image-version"
	)

	labelsAzure := map[string]struct{}{
		labelNameAzureClusterName:      {},
		labelNameAzureOSSKU:            {},
		labelNameAzureRole:             {},
		labelNameAzureNodeImageVersion: {},
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
		if _, ok := labelsAzure[lName]; ok {
			return ClusterProviderAzure, true
		}
	}

	return ClusterProviderUnknown, false
}
