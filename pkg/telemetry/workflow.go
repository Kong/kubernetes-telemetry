package telemetry

import (
	"context"
	"runtime"
	"sync"

	"github.com/gammazero/workerpool"

	"github.com/Kong/kubernetes-telemetry/pkg/provider"
)

// Workflow defines the workflow interface which will be used either for manual
// interaction or in programmed manner in manager.
type Workflow interface {
	// Name returns workflow's name.
	Name() string
	// AddProvider adds a provider.
	AddProvider(provider.Provider)
	// Execute executes the workflow.
	Execute(context.Context) (provider.Report, error)
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
	w.providers = append(w.providers, p)
}

// Execute executes the workflow by triggering all configured providers.
func (w *workflow) Execute(ctx context.Context) (provider.Report, error) {
	var (
		report   = provider.Report{}
		chDone   = make(chan struct{})
		chErr    = make(chan error)
		chReport = make(chan provider.Report)
		wp       = workerpool.New(w.concurrency)
		wg       sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, provider := range w.providers {
		p := provider
		wg.Add(1)

		wp.Submit(func() {
			defer wg.Done()

			report, err := p.Provide(ctx)
			if err != nil {
				chErr <- err
				return
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

forLoop:
	for {
		select {
		case err := <-chErr:
			if err != nil {
				return nil, err
			}
		case r := <-chReport:
			report.Merge(r)
		case <-chDone:
			break forLoop
		}
	}

	return report, nil
}
