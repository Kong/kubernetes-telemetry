package provider

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ErrGVRNotAvailable is an error which indicates that a GVR is not available.
// It contains the reason error which can help find the root cause.
type ErrGVRNotAvailable struct {
	GVR    schema.GroupVersionResource
	Reason error
}

func (e ErrGVRNotAvailable) Error() string {
	return fmt.Sprintf("GVR %q not available, reason: %v", e.GVR, e.Reason)
}
