package telemetry

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
)

func TestWorkflowState(t *testing.T) {
	w, err := NewStateWorkflow()
	require.NoError(t, err)

	r, err := w.Execute(context.Background())
	require.NoError(t, err)

	hostname, err := os.Hostname()
	require.NoError(t, err)

	require.EqualValues(t, provider.Report{
		"uptime": 0,
		"hn":     hostname,
	}, r)
}
