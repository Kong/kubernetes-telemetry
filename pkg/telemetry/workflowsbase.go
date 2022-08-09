package telemetry

import (
	"github.com/kong/kubernetes-telemetry/pkg/provider"
)

const (
	// StateWorkflowName is the name assigned to state workflow.
	StateWorkflowName = "state"
)

// NewStateWorkflow creates a new 'state' workflow, based
// on a predefined set of providers that will deliver telemetry date about the
// state of the system.
func NewStateWorkflow() (Workflow, error) {
	uptimeProvider, err := provider.NewUptimeProvider("uptime")
	if err != nil {
		return nil, err
	}
	hostnameProvider, err := provider.NewHostnameProvider("hostname")
	if err != nil {
		return nil, err
	}

	w := NewWorkflow(StateWorkflowName)
	w.AddProvider(uptimeProvider)
	w.AddProvider(hostnameProvider)

	return w, nil
}
