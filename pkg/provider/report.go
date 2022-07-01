package provider

type Report map[string]any

func (r *Report) Merge(other Report) *Report {
	for k, v := range other {
		(*r)[k] = v
	}
	return r
}
