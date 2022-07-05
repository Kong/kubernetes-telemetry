package provider

import (
	"context"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControllerRuntimeProvideFunc func(ctx context.Context, cl client.Client) (Report, error)

// k8sControllerRuntimeBase is a base boilerplate struct that allows users to create their
// own k8s telemetry providers that interact with the cluster using client.Client.
type k8sControllerRuntimeBase struct {
	// cl provides controller-runtime client implementation.
	cl          client.Client
	provideFunc ControllerRuntimeProvideFunc

	base
}

func NewK8sControllerRuntimeBase(name string, kind Kind, cl client.Client, f ControllerRuntimeProvideFunc) (Provider, error) {
	return k8sControllerRuntimeBase{
		cl:          cl,
		provideFunc: f,
		base: base{
			name: name,
			kind: kind,
		},
	}, nil
}

func (p k8sControllerRuntimeBase) Provide(ctx context.Context) (Report, error) {
	return p.provideFunc(ctx, p.cl)
}

type ClientGoProvideFunc func(ctx context.Context, kc kubernetes.Interface) (Report, error)

// k8sClientGoBase is a base boilerplate struct that allows users to create their
// own k8s telemetry providers that interact with the cluster using kubernetes.Interface.
type k8sClientGoBase struct {
	// kc provides client-go's client implementation. This is used for retrieving
	// cluster's version and architecture through discovery interface.
	kc          kubernetes.Interface
	provideFunc ClientGoProvideFunc

	base
}

func NewK8sClientGoBase(name string, kind Kind, kc kubernetes.Interface, f ClientGoProvideFunc) (Provider, error) {
	return k8sClientGoBase{
		kc:          kc,
		provideFunc: f,
		base: base{
			name: name,
			kind: kind,
		},
	}, nil
}

func (p k8sClientGoBase) Provide(ctx context.Context) (Report, error) {
	return p.provideFunc(ctx, p.kc)
}
