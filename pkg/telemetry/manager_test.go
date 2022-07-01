package telemetry

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	"github.com/Kong/kubernetes-telemetry/pkg/provider"
)

func TestManagerStartStopDoesntFail(t *testing.T) {
	t.Parallel()

	m, err := NewManager(OptManagerLogger(logr.Discard()))
	require.NoError(t, err)
	require.NoError(t, m.Start())
	m.Stop()
}

func TestManagerBasicLogicWorks(t *testing.T) {
	t.Parallel()

	m, err := NewManager(
		OptManagerLogger(logr.Discard()),
		OptManagerPeriod(time.Millisecond),
	)
	require.NoError(t, err)

	{
		w := NewWorkflow("basic1")
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

		m.AddWorkflow(w)
	}

	ch := make(chan provider.Report)
	require.NoError(t, m.AddConsumer(ch))
	require.NoError(t, m.Start())
	require.ErrorIs(t, m.Start(), ErrManagerAlreadyStarted,
		"subsequent starts of the manager should return an error",
	)
	require.ErrorIs(t, m.AddConsumer(make(chan<- provider.Report)),
		ErrCantAddConsumersAfterStart,
		"cannot add consumers after start",
	)

	report := <-ch
	m.Stop()
	require.EqualValues(t, provider.Report{
		"constant1": "value1",
		"constant2": "value2",
	}, report)
}
