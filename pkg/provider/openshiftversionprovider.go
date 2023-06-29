package provider

import (
	"context"

	"github.com/blang/semver/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// OpenShiftVersionKey is a report key used to report the OpenShift version, if any.
	OpenShiftVersionKey = types.ProviderReportKey("openshift_version")
	// OpenShiftVersionKind is the OpenShift cluster version kind.
	OpenShiftVersionKind = Kind(OpenShiftVersionKey)

	// OpenShiftVersionPodNamespace is a namespace expected to contain pods whose environment includes OpenShift version
	// information.
	OpenShiftVersionPodNamespace = "openshift-apiserver-operator"
	// OpenShiftVersionPodApp is a value for the "app" label to select pods whose environment includes OpenShift version
	// information.
	OpenShiftVersionPodApp = "openshift-apiserver-operator"

	// ImageVersionVariable is the environment variable whose value contains the operator image version. For OpenShift
	// operators, this matches the OpenShift version.
	ImageVersionVariable = "OPERATOR_IMAGE_VERSION"
)

// NewOpenShiftVersionProvider provides the OpenShift version, or nothing if the cluster is not OpenShift.
func NewOpenShiftVersionProvider(name string, kc kubernetes.Interface) (Provider, error) {
	return NewK8sClientGoBase(name, OpenShiftVersionKind, kc, openShiftVersionReport)
}

// openShiftVersionReport prepares a report that indicates the OpenShift version. It is empty on non-OpenShift clusters.
func openShiftVersionReport(ctx context.Context, kc kubernetes.Interface) (types.ProviderReport, error) {
	version, found := detectOpenShiftVersion(ctx, kc)
	if !found {
		return types.ProviderReport{}, nil
	}
	return osVersionReport(version), nil
}

// detectOpenShiftVersion checks for the presence of a known OpenShift Pod and obtains the OpenShift version from it.
// It returns the version (which may be empty) and a boolean indicating if the Pod was found at all. If it didn't find
// the Pod, it's probably not OpenShift.
func detectOpenShiftVersion(ctx context.Context, kc kubernetes.Interface) (semver.Version, bool) {
	var (
		pods []corev1.Pod
		cont string
	)
	for {
		list, err := kc.CoreV1().Pods(OpenShiftVersionPodNamespace).List(ctx, metav1.ListOptions{Continue: cont})
		if err != nil {
			return semver.Version{}, false
		}
		pods = append(pods, list.Items...)
		cont = list.Continue
		if cont == "" {
			break
		}
	}
	for _, pod := range pods {
		for _, ev := range pod.Spec.Containers[0].Env {
			if ev.Name == ImageVersionVariable {
				result, err := semver.Parse(ev.Value)
				if err != nil {
					return semver.Version{}, false
				}
				return result, true
			}
		}
	}
	return semver.Version{}, false
}

// osVersionReport transforms a semver Version into a ProviderReport containing an OpenShift version.
func osVersionReport(version semver.Version) types.ProviderReport {
	return types.ProviderReport{
		OpenShiftVersionKey: version.String(),
	}
}
