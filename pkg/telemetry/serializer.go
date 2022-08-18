package telemetry

import "github.com/kong/kubernetes-telemetry/pkg/types"

// Serializer serializes telemetry reports into byte slices.
type Serializer interface {
	Serialize(report types.Report, signal types.Signal) ([]byte, error)
}
