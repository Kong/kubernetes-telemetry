package telemetry

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/puzpuzpuz/xsync"

	"github.com/Kong/kubernetes-telemetry/pkg/provider"
)

const (
	DefaultWorkflowTickPeriod = 5 * time.Second
)

type managerErr string

func (e managerErr) Error() string {
	return string(e)
}

const (
	ErrManagerAlreadyStarted      = managerErr("telemetry manager already started")
	ErrCantAddConsumersAfterStart = managerErr("can't add consumers after start")
)

type manager struct {
	// workflows contains a map of workflows identified by their names
	workflows *xsync.MapOf[Workflow]
	// period defines at what cadence the workflows will be triggered.
	// For now, all workflows work on the same cadence, i.e. are triggered at the
	// same given, ruled by one timer.
	period time.Duration

	// consumers is a slice of channels that will consume reports produced by
	// execution of workflows.
	consumers []chan<- provider.Report

	ch      chan provider.Report
	once    sync.Once
	logger  logr.Logger
	done    chan struct{}
	started int32
}

var _ Manager = (*manager)(nil)

// Manager controls and runs workflows which provide telemetry data.
// This telemetry is then send over to consumers. Owners of consumers are
// responsible that they consume the data in a timely manner.
type Manager interface {
	Start() error
	Stop()
	AddConsumer(ch chan<- provider.Report) error
	AddWorkflow(Workflow)
	Execute(context.Context) (provider.Report, error)
}

func NewManager(opts ...OptManager) (Manager, error) {
	m := &manager{
		workflows: xsync.NewMapOf[Workflow](),
		period:    DefaultWorkflowTickPeriod,
		consumers: []chan<- provider.Report{},
		ch:        make(chan provider.Report),
		logger:    defaultLogger(),
		done:      make(chan struct{}),
	}

	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, fmt.Errorf("failed to create telemetry manager: %w", err)
		}
	}

	return m, nil
}

// AddWorkflow adds a workflow to manager's workflows.
func (m *manager) AddWorkflow(w Workflow) {
	m.workflows.Store(w.Name(), w)
}

// Start starts the manager and periodical workflow execution.
func (m *manager) Start() error {
	if atomic.LoadInt32(&m.started) > 0 {
		return ErrManagerAlreadyStarted
	}

	m.logger.Info("starting telemetry manager")
	go m.workflowsLoop()
	go m.consumerLoop()
	atomic.StoreInt32(&m.started, 1)
	return nil
}

func (m *manager) Stop() {
	m.logger.Info("stopping telemetry manager")
	m.once.Do(func() {
		close(m.done)
	})
}

func (m *manager) AddConsumer(ch chan<- provider.Report) error {
	if atomic.LoadInt32(&m.started) > 0 {
		return ErrCantAddConsumersAfterStart
	}
	m.consumers = append(m.consumers, ch)
	return nil
}

// workflowsLoop defines a mechanism which periodically loops over all configured
// workflows, executes them to get the telemetry data from provided telemetry
// providers and then sends that telemetry over to consumers.
//
// NOTE: for now there is just one period (and hence only one loop) which means
// that all workflow are executed at the same time.
// If there's enough demand in the future this can be done in a way such that each
// workflow has it's own independent period (and hence an independent timelime).
func (m *manager) workflowsLoop() {
	ticker := time.NewTicker(m.period)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return

		case <-ticker.C:
			report, err := m.Execute(context.Background())
			if err != nil {
				m.logger.Error(err, "error executing workflows")
				continue
			}

			select {
			case m.ch <- report:
			case <-m.done:
				break
			}
		}
	}
}

func (m *manager) Execute(ctx context.Context) (provider.Report, error) {
	var (
		err    error
		report = provider.Report{}
	)

	m.workflows.Range(func(name string, v Workflow) bool {
		var r provider.Report
		r, err = v.Execute(ctx)
		if err != nil {
			err = errors.Wrapf(err, "error executing workflow %s", name)
			return false
		}

		report.Merge(r)
		return true
	})
	return report, err
}

// consumerLoop loops over all configured consumers and sends the gathered telemetry
// reports over to them via a channel.
func (m *manager) consumerLoop() {
	for {
		select {
		case <-m.done:
			return

		case r := <-m.ch:
		consumersLoop:
			for _, c := range m.consumers {
				select {
				case c <- r:
				case <-m.done:
					break consumersLoop
				}
			}
		}
	}
}
