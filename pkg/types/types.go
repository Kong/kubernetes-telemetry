package types

// This package was needed to resolve circular dependency.

// Report represents a report that is returned by executing managers workflows.
// This is also the type that is being sent out to consumers.
type Report map[string]ProviderReport

// ProviderReport represents a report that is returned by providers.
type ProviderReport map[ProviderReportKey]any

// ProviderReportKey represents a key type for providers' reports.
type ProviderReportKey string

// Merge merges the report with a different report overriding already existing
// entries if there's a collision.
func (r *ProviderReport) Merge(other ProviderReport) *ProviderReport {
	for k, v := range other {
		(*r)[k] = v
	}
	return r
}

// Signal represents the signal name to include in the serialized report.
type Signal string

// SignalReport contains the packaged Report with a signal name attached.
type SignalReport struct {
	Signal
	Report
}
