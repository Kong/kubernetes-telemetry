package telemetry

import (
	"time"

	"github.com/go-logr/logr"
)

type OptManager func(*manager) error

// OptManagerLogger returns an option that will set manager's logger.
func OptManagerLogger(l logr.Logger) OptManager {
	return func(m *manager) error {
		m.logger = l
		return nil
	}
}

// OptManagerPeriod returns an option that will set manager's workflows period.
func OptManagerPeriod(period time.Duration) OptManager {
	return func(m *manager) error {
		m.period = period
		return nil
	}
}
