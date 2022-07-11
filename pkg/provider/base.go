package provider

import "fmt"

type base struct {
	name string
	kind Kind
}

func (pb base) Name() string {
	return pb.name
}

func (pb base) Kind() Kind {
	return pb.kind
}

// WrapError wraps the error with an error message containing provider's kind and name.
func (pb base) WrapError(err error) error {
	return fmt.Errorf("%s/%s: %w", pb.Kind(), pb.Name(), err)
}
