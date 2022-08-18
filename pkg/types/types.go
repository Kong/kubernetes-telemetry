package types

import "github.com/kong/kubernetes-telemetry/pkg/provider"

// This package was needed to resolve circular dependency.

// Report represents a report that is returned by executing managers workflows.
// This is also the type that is being sent out to consumers.
type Report map[string]provider.Report

// Signal represents the signal name to include in the serialized report.
type Signal string

// SignalReport contains the packaged Report with a signal name attached.
type SignalReport struct {
	Signal
	Report
}
