package telemetry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/puzpuzpuz/xsync/v2"

	"github.com/kong/kubernetes-telemetry/pkg/log"
	"github.com/kong/kubernetes-telemetry/pkg/types"
)

type managerErr string

func (e managerErr) Error() string {
	return string(e)
}

const (
	// ErrManagerAlreadyStarted occurs when a manager has been already started
	// and it's attempted to be started again.
	ErrManagerAlreadyStarted = managerErr("manager already started")
	// ErrCantAddConsumersAfterStart occurs when consumers are tried to be added
	// after the manager has been already started.
	ErrCantAddConsumersAfterStart = managerErr("can't add consumers after start")
	// ErrManagerAlreadyStopped occurs when manager has already been stopped.
	ErrManagerAlreadyStopped = managerErr("manager stopped")
)

const (
	// DefaultWorkflowTickPeriod is the default tick period with which the manager
	// will trigger configured workflows execution.
	DefaultWorkflowTickPeriod = 5 * time.Second
)

type manager struct {
	// signal is the signal that this manager will send out periodically unless
	// overridden via TriggerExecute's parameter.
	signal types.Signal
	// workflows contains a map of workflows identified by their names
	workflows *xsync.MapOf[string, Workflow]
	// period defines at what cadence the workflows will be triggered.
	// For now, all workflows work on the same cadence, i.e. are triggered at the
	// same given, ruled by one timer.
	period time.Duration

	// consumers is a slice of channels that will consume reports produced by
	// execution of workflows.
	consumers []chan<- types.SignalReport

	chTrigger chan types.Signal
	ch        chan types.SignalReport
	once      sync.Once
	logger    logr.Logger
	done      chan struct{}
	started   int32
}

var _ Manager = (*manager)(nil)

// Manager controls and runs workflows which provide telemetry data.
// This telemetry is then send over to consumers. Owners of consumers are
// responsible that they consume the data in a timely manner.
//
// The reports produced by Manager are maps of workflows names - that produced
// their respective reports - to those reports. This way reports from independent
// workflows are enclosed in separate map objects in manager's report.
type Manager interface {
	// Start starts the manager. This in turn starts an internal ticker which
	// periodically triggers the configured workflows to get the telemetry data
	// via the configured providers and to forward that data to consumers.
	Start() error
	// Stop stops the manager the internal loops.
	Stop()
	// AddConsumer adds a consumer of telemetry data provided by configured
	// workflows' providers.
	// AddConsumer(ch chan<- Report) error
	AddConsumer(c Consumer) error
	// AddWorkflow adds a workflow with providers which will provide telemetry data.
	AddWorkflow(Workflow)
	// TriggerExecute triggers an execution of all configured workflows, which will gather
	// all telemetry data, push it downstream to configured serializers and then
	// forward it using the configured forwarders.
	// It will use the provided signal name overriding what's configured in the
	// Manager.
	TriggerExecute(context.Context, types.Signal) error
	// Report executes all workflows and returns an aggregated report from those
	// workflows.
	Report(context.Context) (types.Report, error)
}

// NewManager creates a new manager configured via the provided options.
func NewManager(signal types.Signal, opts ...OptManager) (Manager, error) {
	m := &manager{
		signal:    signal,
		workflows: xsync.NewMapOf[Workflow](),
		period:    DefaultWorkflowTickPeriod,
		consumers: []chan<- types.SignalReport{},
		chTrigger: make(chan types.Signal),
		ch:        make(chan types.SignalReport),
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
	if w == nil {
		return
	}
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

// Stop stops the manager.
func (m *manager) Stop() {
	m.logger.Info("stopping telemetry manager")
	m.once.Do(func() {
		close(m.done)
	})
}

// Consumer is an entity that can consume telemetry reports on a channel returned
// by Intake().
type Consumer interface {
	Intake() chan<- types.SignalReport
	Close()
}

// AddConsumer adds a consumer.
func (m *manager) AddConsumer(c Consumer) error {
	// func (m *manager) AddConsumer(ch chan<- Report) error {
	if atomic.LoadInt32(&m.started) > 0 {
		return ErrCantAddConsumersAfterStart
	}
	m.consumers = append(m.consumers, c.Intake())
	return nil
}

func (m *manager) TriggerExecute(ctx context.Context, signal types.Signal) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.chTrigger <- signal:
		return nil
	case <-m.done:
		return ErrManagerAlreadyStopped
	}
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
	ch := make(chan types.Signal)
	go func() {
		ticker := time.NewTicker(m.period)
		defer ticker.Stop()
		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				ch <- m.signal
			case signal := <-m.chTrigger:
				ch <- signal
			}
		}
	}()

	for {
		select {
		case <-m.done:
			break

		case signal := <-ch:
			ctx, cancel := context.WithTimeout(context.Background(), m.period)

			report, err := m.Report(ctx)
			if err != nil {
				m.logger.V(log.DebugLevel).
					WithValues("error", err.Error()).
					Info("error executing workflows")
			}

			// Continue the execution even if we get an error but account for possibility
			// of getting nil reports, in which case move on to the next iteration (tick).
			if report == nil {
				cancel()
				continue
			}

			select {
			case m.ch <- types.SignalReport{
				Signal: signal,
				Report: report,
			}:
			case <-m.done:
				cancel()
				break
			}
			cancel()
		}
	}
}

// Execute executes all configures workflows and returns an aggregated report
// from all the underlying providers.
func (m *manager) Report(ctx context.Context) (types.Report, error) {
	var (
		errs   []error
		report = types.Report{}
	)

	m.workflows.Range(func(name string, w Workflow) bool {
		r, err := w.Execute(ctx)
		if err != nil {
			errs = append(errs, err)
		}

		// Add the report regardless if it's partial only omitting empty (nil) reports.
		if r != nil {
			report[w.Name()] = r
		}

		return true
	})
	return report, errors.Join(errs...)
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
