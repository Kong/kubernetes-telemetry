package telemetry

import (
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
)

const (
	// MeshDetectWorkflowName is the name assigned to mesh detect workflow.
	MeshDetectWorkflowName = "mesh-detect"
)

// NewMeshDetectWorkflow returns a mesh detection workflow.
//
// Exemplar report produced:
//
//	{
//		"mdist": "all8,a2,c2",
//		"mdep": "a3,c3"
//	}
func NewMeshDetectWorkflow(cl client.Client, pod, publishService apitypes.NamespacedName) (Workflow, error) {
	if cl == nil {
		return nil, ErrNilControllerRuntimeClientProvided
	}

	pMeshDetect, err := provider.NewMeshDetectProvider(string(provider.MeshDetectProviderKey), cl, pod, publishService)
	if err != nil {
		return nil, err
	}

	w := NewWorkflow(MeshDetectWorkflowName)
	w.AddProvider(pMeshDetect)

	return w, nil
}
