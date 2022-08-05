package telemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
)

func TestWorkflow(t *testing.T) {
	w := NewWorkflow("test1")

	{
		p, err := provider.NewFixedValueProvider("constant1", provider.Report{
			"constant1": "value1",
		})
		require.NoError(t, err)
		w.AddProvider(p)
	}
	{
		p, err := provider.NewFixedValueProvider("constant2", provider.Report{
			"constant2": "value2",
		})
		require.NoError(t, err)
		w.AddProvider(p)
	}

	report, err := w.Execute(context.Background())
	require.NoError(t, err)

	require.EqualValues(t, provider.Report{
		"constant1": "value1",
		"constant2": "value2",
	}, report)
}
