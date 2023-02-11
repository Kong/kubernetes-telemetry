package telemetry

import (
	"context"
	"runtime"
	"sync"

	"github.com/gammazero/workerpool"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/kong/kubernetes-telemetry/pkg/provider"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

// Workflow defines the workflow interface which will be used either for manual
// interaction or in programmed manner in manager.
type Workflow interface {
	// Name returns workflow's name.
	Name() string
	// AddProvider adds a provider.
	AddProvider(provider.Provider)
	// Execute executes the workflow.
	Execute(context.Context) (types.ProviderReport, error)
}

var _ Workflow = (*workflow)(nil)

type workflow struct {
	name        string
	concurrency int
	providers   []provider.Provider
}

// NewWorkflow creates a new empty workflow.
func NewWorkflow(name string) Workflow {
	return &workflow{
		name:        name,
		concurrency: runtime.NumCPU(),
		providers:   make([]provider.Provider, 0),
	}
}

// Name returns workflow's name.
func (w *workflow) Name() string {
	return w.name
}

// AddProvider adds provider to the list of configured providers.
func (w *workflow) AddProvider(p provider.Provider) {
	if p == nil {
		return
	}
	w.providers = append(w.providers, p)
}

// Execute executes the workflow by triggering all configured providers.
func (w *workflow) Execute(ctx context.Context) (types.ProviderReport, error) {
	var (
		report   = types.ProviderReport{}
		chDone   = make(chan struct{})
		chErr    = make(chan error)
		chReport = make(chan types.ProviderReport)
		wp       = workerpool.New(w.concurrency)
		wg       sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(len(w.providers))
	for _, provider := range w.providers {
		p := provider
		wp.Submit(func() {
			defer wg.Done()

			report, err := p.Provide(ctx)
			if err != nil {
				chErr <- errors.Wrapf(err, "problem with provider %s", p.Name())
			}

			chReport <- report
		})
	}

	go func() {
		wg.Wait()
		close(chDone)
		close(chErr)
		close(chReport)
	}()

	var mErr error

forLoop:
	for {
		select {
		case err := <-chErr:
			if err != nil {
				mErr = multierror.Append(mErr,
					errors.Wrapf(err, "error executing workflow %s", w.Name()),
				)
			}
		case r := <-chReport:
			report.Merge(r)
		case <-chDone:
			break forLoop
		}
	}

	return report, mErr
}
