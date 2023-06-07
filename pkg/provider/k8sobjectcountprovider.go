package provider

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

// k8sObjectCount is a Provider that returns the count of all objects of a certain kind in the Kubernetes cluster.
// Caller indicates object kind by passing its GroupVersionResource to `objectType`.
//
// Example: Use {Group: "", Version: "v1", Resource: "pods"} to get a Provider that counts all Pods in the cluster.
type k8sObjectCount struct {
	gvr      schema.GroupVersionResource
	resource dynamic.NamespaceableResourceInterface

	base
}

// NewK8sObjectCountProvider returns a k8s object count provider, which will provide a count of
// specified resource.
func NewK8sObjectCountProvider(name string, kind Kind, d dynamic.Interface, gvr schema.GroupVersionResource) (Provider, error) {
	p := &k8sObjectCount{
		resource: d.Resource(gvr),
		gvr:      gvr,
		base: base{
			name: name,
			kind: kind,
		},
	}

	return p, nil
}

// NewK8sObjectCountProviderWithRESTMapper returns a k8s object count provider and it will use the
// provided rest mapper to check if there is a kind for the provided group version resource,
// available on the cluster.
func NewK8sObjectCountProviderWithRESTMapper(name string, kind Kind, d dynamic.Interface, gvr schema.GroupVersionResource, rm meta.RESTMapper) (Provider, error) {
	p, err := NewK8sObjectCountProvider(name, kind, d, gvr)
	if err != nil {
		return nil, err
	}

	if err := p.(*k8sObjectCount).GVRInCluster(rm); err != nil {
		// If there's no kind for object in the cluster then just don't add its count provider.
		if meta.IsNoMatchError(err) {
			return nil, ErrGVRNotAvailable{
				GVR:    gvr,
				Reason: err,
			}
		}
		if errG := (&discovery.ErrGroupDiscoveryFailed{}); errors.As(err, &errG) {
			return nil, ErrGVRNotAvailable{
				GVR:    gvr,
				Reason: errG,
			}
		}
		return nil, err
	}

	return p, nil
}

const (
	defaultPageLimit = 1000
)

func (k *k8sObjectCount) Provide(ctx context.Context) (types.ProviderReport, error) {
	var (
		count         int
		continueToken string
	)

	for {
		list, err := k.resource.List(ctx, metav1.ListOptions{
			// Conservatively use a limit for paging.
			Limit:    defaultPageLimit,
			Continue: continueToken,
		})
		if err != nil {
			return types.ProviderReport{}, k.WrapError(err)
		}

		count += len(list.Items)
		if continueToken = list.GetContinue(); continueToken == "" {
			break
		}
	}

	return types.ProviderReport{
		types.ProviderReportKey("k8s_" + k.gvr.Resource + "_count"): count,
	}, nil
}

func (k *k8sObjectCount) GVRInCluster(rm meta.RESTMapper) error {
	_, err := rm.KindFor(k.gvr)
	return err
}
