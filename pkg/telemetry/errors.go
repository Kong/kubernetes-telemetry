package telemetry

type err string

func (e err) Error() string {
	return string(e)
}

const (
	// ErrNilRestConfig occurs when a nil *rest.Config is provided.
	ErrNilRestConfig = err("provided nil *rest.Config")
	// ErrNilDynClientProvided occurs when a nil dynamic.Interface is provided.
	ErrNilDynClientProvided = err("provided nil dynamic.Interface")
	// ErrNilKubernetesInterfaceProvided occurs when a nil kubernetes.Interface is provided.
	ErrNilKubernetesInterfaceProvided = err("provided nil kubernetes.Interface")
)
