package meshdetect

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

// DeploymentResults is the result of detecting signals of whether a certain
// mesh is deployed in kubernetes cluster.
type DeploymentResults struct {
	ServiceExists bool `json:"serviceExists"`
}

// RunUnderResults is the result of detecting signals of whether pod is
// running under a certain service mesh.
type RunUnderResults struct {
	PodOrServiceAnnotation   bool `json:"podOrServiceAnnotation"`
	SidecarContainerInjected bool `json:"sidecarContainerInjected"`
	InitContainerInjected    bool `json:"initContainerInjected"`
}

// ServiceDistributionResults contains number of total services and number of
// services running under each mesh.
type ServiceDistributionResults struct {
	TotalServices int `json:"totalServices"`
	// MeshDistribution is the number of services running under each kind of mesh.
	// We decided to directly use number here instead of ratio in total services in a
	// floating number because using floating number needs extra work on calculating
	// and serializing.
	MeshDistribution map[MeshKind]int `json:"meshDistribution"`
}

func (s *ServiceDistributionResults) ToProviderReport() types.ProviderReport {
	if s == nil {
		return nil
	}

	// format: mdist="all100,a10,i20,k50,km50"

	value := "all" + strconv.Itoa(s.TotalServices)

	if s.MeshDistribution != nil {
		// append number of services running in the mesh, if there are any.
		var signals []string
		for _, meshKind := range MeshesToDetect {
			num := s.MeshDistribution[meshKind]
			if num > 0 {
				signals = append(signals, meshKindShortNames[meshKind]+strconv.Itoa(num))
			}
		}

		if len(signals) > 0 {
			// sort the signals to produce a constistent output.
			sort.Strings(signals)
			value = value + "," + strings.Join(signals, ",")
		}
	}

	return types.ProviderReport{
		"mdist": value,
	}
}
