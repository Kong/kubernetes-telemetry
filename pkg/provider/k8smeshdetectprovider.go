package provider

import (
	"context"
	"fmt"

	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kubernetes-telemetry/pkg/internal/meshdetect"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

const (
	// MeshDetectProviderKey is report key under which the mesh detectino info
	// will be provided.
	MeshDetectProviderKey = types.ProviderReportKey("mesh_detect")
	// MeshDetectKind represents the mesh detect provider kind.
	MeshDetectKind = Kind("mesh_detect")
)

// NewMeshDetectProvider returns a mesh detection provider, which will provide
// a information about detected meshes in kubernetes cluster.
func NewMeshDetectProvider(name string, cl client.Client, pod, publishService apitypes.NamespacedName) (Provider, error) {
	d, err := meshdetect.NewDetectorByConfig(cl, pod, publishService)
	if err != nil {
		return nil, err
	}

	p := newMeshDetectProvider(name, d)

	return p, nil
}

func newMeshDetectProvider(name string, d *meshdetect.Detector) Provider {
	return meshDetectProvider{
		detector: d,
		base: base{
			name: name,
			kind: MeshDetectKind,
		},
	}
}

type meshDetectProvider struct {
	detector *meshdetect.Detector
	base
}

func (md meshDetectProvider) Provide(ctx context.Context) (types.ProviderReport, error) {
	r := make(types.ProviderReport)

	deploymentResults := md.detector.DetectMeshDeployment(ctx)
	r.Merge(deploymentResults.ToProviderReport())

	runUnderResults, err := md.detector.DetectRunUnder(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect pod's mesh: %w", err)
	}
	r.Merge(runUnderResults.ToProviderReport())

	serviceDistributionResults, err := md.detector.DetectServiceDistribution(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to detect service distribution under meshes: %w", err)
	}
	r.Merge(serviceDistributionResults.ToProviderReport())

	return r, nil
}
