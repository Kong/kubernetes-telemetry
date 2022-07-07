package provider

// Report represents a report from a provider.
type Report map[string]any

// Merge merges the report with a different report overriding already existing
// entries if there's a collision.
func (r *Report) Merge(other Report) *Report {
	for k, v := range other {
		(*r)[k] = v
	}
	return r
}
