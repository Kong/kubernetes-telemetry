---
title: Methods to Detect Service Mesh Deployment and Distribution
status: provisional
authors: randmonkey
creation-date: 2022-05-30
---

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
- [Proposal](#proposal)
<!-- /toc -->

## Summary 

[Service Mesh][servicemesh] is an infrastructure layer that allows you to 
manage communication between your application’s microservices. It is
important for us to know which service mesh our customers and end users 
uses in their kubernetes clusters, and how do they use service mesh.
Currently, the usage of service mesh is not collected in our products. The
purpose of this KEP is to implement a method to detect that which kind of 
(and whether) service mesh is deployed in kubernetes cluster, and the rate of
services in the mesh, which would be used in the
[Kong Kubernetes Ingress Controller (KIC)][kic] or other products written in
[Golang][go] running in kubernetes.

[servicemesh]:https://www.digitalocean.com/community/tutorials/an-introduction-to-service-meshes
[kic]:https://github.com/kong/kubernetes-ingress-controller
[go]:https://go.dev

## Motivation

- We want to detect if our users deploys service mesh in their kubernetes\
  clusters, and which service mesh do they deploy (for example, istio,
  linkerd, or kuma)
- We want to detect if KIC[kic] and Kong Gateway[konggw] is running in a
  service mesh.
- We want to detect the share of services running in service mesh.
- We want to implement the detection methods above in the telemetry
  library for products such as KIC[kic] to import.

[kic]:https://github.com/kong/kubernetes-ingress-controller
[konggw]:https://docs.konghq.com/gateway/

### Goals
- implement detection method as a golang library for the follwing 7 kinds of
  service meshes: `istio`,`linkerd`,`kuma`,`kong mesh`,`consul connect`,
  `traefik mesh`,`aws app mesh`.
- define the format of reporting the results of service mesh detection in 
  [splunk](splunk) format.

[splunk]:https://github.com/splunk

## Proposal

### Detecting Service Mesh Deployment

This part provides methods to detect if a service mesh is deployed, and what
kind of service mesh is deployed. Generally, there are 3 signals can be used
to show that a service mesh is deployed. They are:

- Presence of namespaces where components of service meshes are in (e.g.: 
  `istio-system`,`kuma-system`)
- Presence of CRDs used by service mesh (e.g.: 
  `virtualservices.networking.istio.io`,`meshes.kuma.io`).
- Existence of `services` of service mesh components in certain namespace or
  `default` namespace (e.g.:`istiod`,`kuma-control-plane`)

Here is a table on the details on how to detect whether a specific method of
service mesh.

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
|signal1: namespace	|`istio-system` 	|`linkerd` |`kuma-system`   	| `kong-mesh-system`  	|`consul`| `traefik-mesh`|`appmesh-system`|
|  signal2: CRD(one example)|   `virtualservices.networking.istio.io`	|   `serviceprofiles.linkerd.io`	|   `meshes.kuma.io`	|(same as kuma)|`meshes.consul.hashicorp.com`|(NO unique CRDs)|`meshes.appmesh.k8s.aws`|
|   signal3: service(one example) 	|   `istiod`	| `linkerd-proxy-injector`  	|   `kuma-control-plane`	| `kong-mesh-control-plane` | `consul-server` | `traefik-mesh-controller` | `appmesh-controller-webhook-service` |

### Detecting KIC Running in Service Mesh

This part provides methods to detect if [KIC][kic] running in a service mesh.
We have 4 different signals to know if KIC is running in a service mesh:

- namespace where KIC in has certain annotations or labels(e.g.: label 
`istio-injection=enabled`, annotation `kuma.io/sidecar-injection=enabled`)
- annotations or labels in KIC pod or service(e.g.: annotation
`sidecar.istio.io/status`,`kuma.io/sidecar-injected=true` in KIC pod)
- sidecar container in KIC pod
- initContainer injected by service mesh in KIC pod

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
|signal1: namespace annotation/label| label `istio-injection=enabled` | annotation `linkerd.io/inject=enabled` |annotation `kuma.io/sidecar-injection=enabled`|same as kuma | (NO such annotation/label) | (NO such annotation/label) | `appmesh.k8s.aws/sidecarInjectorWebhook=enabled`|
|signal2: pod/service annotations | pod: `sidecar.istio.io/status` | pod: `linkerd.io/proxy-version` | pod: `kuma.io/sidecar-injected=true` | same as kuma | pod: `consul.hashicorp.com/connect-inject-status=injected` | service: `mesh.traefik.io/traffic-type:(HTTP|TCP)` | NO certain annotations | 
| signal3: sidecar container| `istio-proxy` | `linkerd-proxy`  | `kuma-sidecar` | same as kuma | `envoy-sidecar` | NO sidecar | `envoy` | 
| signal4: init container | `istio-init` | `linkerd-init` | `kuma-init` | same as kuma | `consul-connect-inject-init`  | NO init container | `proxyinit` | 

### Detecting Distribution of Service Mesh

This part proposes methods to detect the rate of services running in service
mesh. Usually, pods of services running in service mesh are injected with a 
sidecar container. We can list all `services` in kubernetes cluster, and then
check their correspoding `endpoints` to get the names of `pods` in the
service. Finally, we check whether the `pods` has a sidecar container to
judge whether the service is running in service mesh.

For example, we check all `services` in `default` namespace:

```bash
$ kubectl get services
NAME                 TYPE        CLUSTER-IP       EXTERNAL-IP   PORT(S)           AGE
httpbin-deployment   ClusterIP   10.152.183.181   <none>        80/TCP            20d
echo                 ClusterIP   10.152.183.89    <none>        8080/TCP,80/TCP   19d
kubernetes           ClusterIP   10.152.183.1     <none>        443/TCP           40d
``` 

Then, we check the `subsets` field of `endpoints/echo` to get names of pods:

```bash
$ kubectl get endpoints echo -o jsonpath={.subsets} | json_pp
```

result would be like

```json
[
   {
      "addresses" : [
         {
            "ip" : "10.1.219.40",
            "nodeName" : "node-name",
            "targetRef" : {
               "kind" : "Pod",
               "name" : "echo-588c888c78-w5nsn",
               "namespace" : "default",
               "resourceVersion" : "1307822",
               "uid" : "ae5ac774-cd72-410a-9f30-7ceea42f0498"
            }
         }
      ],
      "ports" : [
          // ...
      ]
   }
]
```

The results of the `targetRef.name` showed the name of pods in services, and
then we get the `spec.containers[*].name` field of pods:

```bash
$ kubectl get pods echo-588c888c78-w5nsn -ojsonpath={.spec.containers[*].name}
echo istio-proxy
```

We can judge that service `default/echo` is running in istio mesh, since
there is a `istio-proxy` sidecar container in its pods.

However, traefik mesh is an exception of the method above since it does not
inject sidecar containers. But the method for judging whether a service is 
running in traefik mesh is much simpler. We can check if `service` has one 
of the following annotation: `mesh.traefik.io/traffic-type:HTTP` 
or `mesh.traefik.io/traffic-type:TCP`.
