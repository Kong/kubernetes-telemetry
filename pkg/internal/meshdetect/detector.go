package meshdetect

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	apitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kubernetes-telemetry/pkg/types"
)

// Detector provides methods to detect the following:
//
//   - whether a service mesh is deployed to the cluster
//   - whether the pod is injected with service mesh
//   - count of services injected within mesh networks
type Detector struct {
	// Client is the kubernetes client to read kubernetes services.
	Client client.Client

	// Pod is the pod in which the mesh detector is running.
	Pod apitypes.NamespacedName

	// PublishServiceName is the Kubernetes service used for ingress traffic
	// to the Kong Gateway.
	PublishService apitypes.NamespacedName
}

const (
	// defaultPageSize is the default limit of each single call of
	// listing all resources(services,endpoints,pods) in pages.
	defaultPageSize = 1000
)

// NewDetectorByConfig creates a new Detector provided a Kubernetes
// config for the relevant cluster and the name of the Kubernetes service
// for ingress traffic to the Kong Gateway.
func NewDetectorByConfig(
	client client.Client,
	pod apitypes.NamespacedName,
	publishService apitypes.NamespacedName,
) (*Detector, error) {
	return &Detector{
		Client:         client,
		Pod:            pod,
		PublishService: publishService,
	}, nil
}

type meshDeploymentResults map[MeshKind]*DeploymentResults

func (m meshDeploymentResults) ToProviderReport() types.ProviderReport {
	if len(m) == 0 {
		return nil
	}

	signals := []string{}
	for _, meshKind := range MeshesToDetect {
		result := m[meshKind]
		if result == nil {
			continue
		}
		// signal3: service exists
		if result.ServiceExists {
			signals = append(signals, meshKindShortNames[meshKind]+"3")
		}
	}

	if len(signals) == 0 {
		return nil
	}

	// sort the signals (in alphabetical order),
	// then join them together to produce a consistent output for same results.
	sort.Strings(signals)

	return types.ProviderReport{
		"mdep": strings.Join(signals, ","),
	}
}

// DetectMeshDeployment detects which kinds of mesh networks are deployed.
func (d *Detector) DetectMeshDeployment(ctx context.Context) meshDeploymentResults {
	deploymentResults := meshDeploymentResults{}

	for _, meshKind := range MeshesToDetect {
		deploymentResult := &DeploymentResults{}
		if d.detectMeshDeploymentByService(ctx, meshKind) {
			deploymentResult.ServiceExists = true
		}
		deploymentResults[meshKind] = deploymentResult
	}

	return deploymentResults
}

// detectMeshDeploymentByService finds the service for each mesh in the cluster.
func (d *Detector) detectMeshDeploymentByService(ctx context.Context, meshKind MeshKind) bool {
	serviceName := meshServiceName[meshKind]

	svcList := &corev1.ServiceList{}
	err := d.Client.List(ctx, svcList, &client.ListOptions{
		FieldSelector: fields.SelectorFromSet(fields.Set{"metadata.name": serviceName}),
	})
	if err != nil {
		return false
	}

	for _, svc := range svcList.Items {
		if svc.Name == serviceName {
			return true
		}
	}

	return false
}

type meshRunUnderResults map[MeshKind]*RunUnderResults

// DetectRunUnder detects whether the pod is running under each kind of service mesh.
// in this function, we want to detect whether the pod running this detector, which is
// also the pod, is running under a certain kind of service mesh.
// for example, if the pod is injected with istio sidecar container and init container,
// we report that the detector is running under istio mesh.
func (d *Detector) DetectRunUnder(ctx context.Context) (meshRunUnderResults, error) {
	runUnderResults := meshRunUnderResults{}
	// get the pod itself.
	pod := &corev1.Pod{}
	err := d.Client.Get(ctx, client.ObjectKey{Namespace: d.Pod.Namespace, Name: d.Pod.Name}, pod)
	if err != nil {
		return nil, fmt.Errorf("failed to get current pod %v: %w", d.Pod, err)
	}

	publishService := &corev1.Service{}
	// only try to get service if the namespace and name are correctly filled
	if d.PublishService.Name != "" && d.PublishService.Namespace != "" {
		err := d.Client.Get(ctx, d.PublishService, publishService)
		if err != nil {
			return nil, fmt.Errorf("failed to get publish service %v: %w", d.PublishService, err)
		}
	}

	for _, meshKind := range MeshesToDetect {
		runUnderResults[meshKind] = &RunUnderResults{}

		// detect if service for kong-gateway has annotations(only for traefik)
		if publishService != nil && isServiceInjected(meshKind, publishService) {
			runUnderResults[meshKind].PodOrServiceAnnotation = true
		}

		// detect if pod has annotations.
		podAnnotations := meshPodAnnotations[meshKind]
		if podAnnotations != nil && podAnnotations.Matches(labels.Set(pod.Annotations)) {
			runUnderResults[meshKind].PodOrServiceAnnotation = true
		}

		// detect if pod has a sidecar.
		runUnderResults[meshKind].SidecarContainerInjected = isPodSidecarInjected(meshKind, pod)

		// detect if pod has a init container.
		runUnderResults[meshKind].InitContainerInjected = isPodInitContainerInjected(meshKind, pod)
	}

	return runUnderResults, nil
}

func (m meshRunUnderResults) ToProviderReport() types.ProviderReport {
	if len(m) == 0 {
		return nil
	}

	signals := []string{}
	for _, meshKind := range MeshesToDetect {
		result := m[meshKind]
		if result == nil {
			continue
		}

		// signal2: pod/service has annotation
		if result.PodOrServiceAnnotation {
			signals = append(signals, meshKindShortNames[meshKind]+"2")
		}
		// signal3: sidecar injected
		if result.SidecarContainerInjected {
			signals = append(signals, meshKindShortNames[meshKind]+"3")
		}
		// signal4: init container injected
		if result.InitContainerInjected {
			signals = append(signals, meshKindShortNames[meshKind]+"4")
		}
	}

	if len(signals) == 0 {
		return nil
	}
	// sort the signals to produce a constistent output.
	sort.Strings(signals)
	value := strings.Join(signals, ",")
	return types.ProviderReport{
		"kinm": value,
	}
}

func isServiceInjected(meshKind MeshKind, svc *corev1.Service) bool {
	if svc == nil {
		return false
	}
	if svc.Annotations == nil {
		return false
	}

	svcAnnotations := meshServiceAnnotations[meshKind]
	if svcAnnotations == nil {
		return false
	}
	if svcAnnotations.Matches(labels.Set(svc.Annotations)) {
		return true
	}

	return false
}

func isPodSidecarInjected(meshKind MeshKind, pod *corev1.Pod) bool {
	sidecarName := meshSidecarContainerName[meshKind]
	if sidecarName == "" {
		return false
	}
	for _, container := range pod.Spec.Containers {
		if container.Name == sidecarName {
			switch meshKind {
			case MeshKindAWSAppMesh:
				// special judgement for AWS appmesh here:
				// AWS appmesh uses `envoy` as sidecar name, which is a very common name.
				// We do a further check on the container and treat as really injected
				// if the container uses image `aws-appmesh-envoy:*`.
				if strings.Contains(container.Image, awsAppMeshEnvoyImageName) {
					return true
				}
			default:
				// for meshes other than AWS app mesh, directly return true (pod injected)
				// when the container with the sidecar name is found.
				return true
			}
		}
	}
	return false
}

func isPodInitContainerInjected(meshKind MeshKind, pod *corev1.Pod) bool {
	initContainerName := meshInitContainerName[meshKind]
	if initContainerName == "" {
		return false
	}
	for _, initContainer := range pod.Spec.InitContainers {
		if initContainer.Name == initContainerName {
			return true
		}
	}

	return false
}

// listAllServices returns all services in all namespaces.
func (d *Detector) listAllServices(ctx context.Context, pageSize int) ([]*corev1.Service, error) {
	serviceList := []*corev1.Service{}
	continueToken := ""
	for {
		services := &corev1.ServiceList{}
		err := d.Client.List(ctx, services, client.Limit(pageSize), client.Continue(continueToken))
		if err != nil {
			return nil, err
		}

		for i := range services.Items {
			serviceList = append(serviceList, &services.Items[i])
		}

		continueToken = services.GetContinue()

		if services.RemainingItemCount == nil || *services.RemainingItemCount == 0 {
			break
		}
	}

	return serviceList, nil
}

// listAllEndpointsPerService lists all Endpoints per every Service in every namespaces.
// returns map: namespaced name of Service -> slice of Endpoint resources
//
// example: client.ObjectKey{Namespace: "default", Name: "service1"} ->
//
//	[]discoveryv1.Endpoint{
//		{ TargetRef: &v1.ObjectReference{Kind: "Pod", Namespace: "default", Name: "pod1"...}... },
//	    { TargetRef: &v1.ObjectReference{Kind: "Pod", Namespace: "default", Name: "pod2"...}... },
//	    ...
//	}.
func (d *Detector) listAllEndpointsPerService(ctx context.Context, pageSize int) (
	map[client.ObjectKey][]discoveryv1.Endpoint, error,
) {
	endpointsMap := map[client.ObjectKey][]discoveryv1.Endpoint{}
	var continueToken string
	for {
		endpoints := &discoveryv1.EndpointSliceList{}
		if err := d.Client.List(
			ctx, endpoints, client.Limit(pageSize), client.Continue(continueToken),
		); err != nil {
			return nil, err
		}
		for i := range endpoints.Items {
			es := &endpoints.Items[i]
			svcName := es.Labels[discoveryv1.LabelServiceName]
			if svcName == "" {
				continue
			}
			key := client.ObjectKey{Namespace: es.Namespace, Name: svcName}
			endpointsMap[key] = append(endpointsMap[key], es.Endpoints...)
		}

		continueToken = endpoints.GetContinue()
		if endpoints.RemainingItemCount == nil || *endpoints.RemainingItemCount == 0 {
			break
		}
	}
	return endpointsMap, nil
}

// listAllPods lists all pods in all namespaces.
// returns map: namespaced name of pod -> pod resource
// example: client.ObjectKey{Namespace: "default", Name: "pod1"} ->
//
//	&corev1.Pod{
//			ObjectMeta: metav1.ObjectMeta {
//				Namespace: "default",
//				Name: "service1", ...
//			},
//			Spec: ...,
//			...
//	}.
func (d *Detector) listAllPods(ctx context.Context, pageSize int) (
	map[client.ObjectKey]*corev1.Pod, error,
) {
	podMap := map[client.ObjectKey]*corev1.Pod{}
	continueToken := ""
	for {
		partialPodList := &corev1.PodList{}
		err := d.Client.List(ctx, partialPodList, client.Limit(pageSize), client.Continue(continueToken))
		if err != nil {
			return nil, err
		}

		for i := range partialPodList.Items {
			pod := &partialPodList.Items[i]
			key := client.ObjectKey{Namespace: pod.Namespace, Name: pod.Name}
			podMap[key] = pod
		}

		continueToken = partialPodList.GetContinue()
		if partialPodList.RemainingItemCount == nil || *partialPodList.RemainingItemCount == 0 {
			break
		}
	}
	return podMap, nil
}

// DetectServiceDistribution detects how many services are running under each mesh.
func (d *Detector) DetectServiceDistribution(ctx context.Context) (*ServiceDistributionResults, error) {
	// list all services, endpoints and pods to check whether
	// pods behind each service is injected by each service mesh.

	serviceList, err := d.listAllServices(ctx, defaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list services in cluster: %w", err)
	}

	endpointsPerSvc, err := d.listAllEndpointsPerService(ctx, defaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list endpoints in cluster: %w", err)
	}

	pods, err := d.listAllPods(ctx, defaultPageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in cluster: %w", err)
	}

	ret := &ServiceDistributionResults{
		MeshDistribution: map[MeshKind]int{},
		TotalServices:    len(serviceList),
	}

	for _, svc := range serviceList {
		endpoints := endpointsPerSvc[client.ObjectKeyFromObject(svc)]
		if endpoints == nil {
			continue
		}

		// Injected is set to true if the service(pod) is injected by mesh.
		injected := map[MeshKind]bool{}

		// Detect if service has annotations to indicate that the service is injected
		// (only for traefik).
		for meshKind := range meshServiceAnnotations {
			injected[meshKind] = isServiceInjected(meshKind, svc)
		}

		for _, endpoint := range endpoints {
			if endpoint.TargetRef == nil || endpoint.TargetRef.Kind != "Pod" {
				continue
			}
			// If one of the Pods is injected, we consider this service as running under the mesh.
			podKey := client.ObjectKey{
				Namespace: endpoint.TargetRef.Namespace,
				Name:      endpoint.TargetRef.Name,
			}
			pod := pods[podKey]
			if pod == nil {
				continue
			}

			for _, meshKind := range MeshesToDetect {
				// Set injected to true if one of pods in service is injected with sidecar and init container.
				injected[meshKind] = injected[meshKind] ||
					isPodSidecarInjected(meshKind, pod) ||
					isPodInitContainerInjected(meshKind, pod)
			}
		}

		for meshKind := range injected {
			if injected[meshKind] {
				ret.MeshDistribution[meshKind]++
			}
		}
	}

	return ret, nil
}
