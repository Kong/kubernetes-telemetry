package provider

import "fmt"

type base struct {
	kind, name string
}

func (pb base) Name() string {
	return pb.name
}

func (pb base) Kind() string {
	return pb.kind
}

func (pb base) WrapError(err error) error {
	return fmt.Errorf("%s/%s: %w", pb.Kind(), pb.Name(), err)
}
