package types

import "github.com/kong/kubernetes-telemetry/pkg/provider"

// This package was needed to resolve circular dependency.

// Report represents a report that is returned by executing managers workflows.
// This is also the type that is being sent out to consumers.
type Report map[string]provider.Report
