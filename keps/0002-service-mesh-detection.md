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
use in their kubernetes clusters (and how they use those service meshes)
so that we can predict and plan for performance tunings, features and new
products which integrate with mesh networks. Currently, the usage of service
mesh is not collected in our products. The purpose of this KEP is to
implement a method to detect whether a service mesh is deployed in kubernetes
cluster and if so gather information about this mesh, This is in turn could
be used in the  [Kong Kubernetes Ingress Controller (KIC)][kic] or other
products written in [Golang][go] running in kubernetes.

[servicemesh]:https://www.digitalocean.com/community/tutorials/an-introduction-to-service-meshes
[kic]:https://github.com/kong/kubernetes-ingress-controller
[go]:https://go.dev

## Motivation

- We want to detect if our users deploy service mesh in their kubernetes
  clusters, and which service mesh do they deploy (for example, istio,
  linkerd, or kuma)
- We want to detect if [KIC][kic] and/or [Kong Gateway][konggw] is running
  within a service mesh network.
- We want to detect the number of services running in service mesh.
- We want to implement the detection methods above in the telemetry
  library for products such as [KIC][kic] to import.

[kic]:https://github.com/kong/kubernetes-ingress-controller
[konggw]:https://docs.konghq.com/gateway/

### Goals
- Implement detection method as a golang library for the follwing 7 kinds of
  service meshes: `istio`,`linkerd`,`kuma`,`kong mesh`,`consul connect`,
  `traefik mesh`,`aws app mesh`.
- Define the format of reporting the results of service mesh detection to 
  our [splunk][splunk] service.

[splunk]:https://github.com/splunk

## Proposal

### Detecting Service Mesh Deployment

This part provides methods to detect if a service mesh is deployed, and what
kind of service mesh is deployed. Generally, there are 3 signals that can be 
used to show that a service mesh is deployed. They are:

- Presence of namespaces where components of service meshes are in (e.g.: 
  `istio-system`,`kuma-system`). This method is not so accurate, which may
  have two misleading situations: (1) service mesh deployed in different 
  namespace (2) namespace remains after a service mesh uninstalled.
- Presence of CRDs used by service mesh (e.g.: 
  `virtualservices.networking.istio.io`,`meshes.kuma.io`). This may also be
  inaccurate, since there would be CRDs remaining after an imcomplete
  uninstallment of a service mesh.
- Existence of `services` of service mesh components in certain namespace or
  `default` namespace (e.g.:`istiod`,`kuma-control-plane`). This could be
  more accurate, since the existence of services indicates that components of
  the service mesh is deployed.

Here is a table on the details on how to detect whether a specific method of
service mesh.

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
|signal1: namespace	|`istio-system` 	|`linkerd` |`kuma-system`  	| `kong-mesh-system`  	|`consul`| `traefik-mesh`|`appmesh-system`|
|  signal2: CRD(one example)|   `virtualservices.networking.istio.io`	|   `serviceprofiles.linkerd.io`	|   `meshes.kuma.io`	|(same as kuma)|`meshes.consul.hashicorp.com`|(NO unique CRDs)|`meshes.appmesh.k8s.aws`|
|   signal3: service(one example) 	|   `istiod`	| `linkerd-proxy-injector`  	|   `kuma-control-plane`	| `kong-mesh-control-plane` | `consul-server` | `traefik-mesh-controller` | `appmesh-controller-webhook-service` |

Special notes:

- All the service meshes could be installed in a namespace other than the
  namespace in the table.
- Namespace for kuma/kong mesh has label `kuma.io/system-namespace: "true"`
- kuma and kong mesh could be distinguished by namespace name and service 
  name.
- `traefik mesh` does not have unique CRDs.

### Detecting KIC Running in Service Mesh

This part provides methods to detect if [KIC][kic] running in a service mesh.
We have 4 different signals to know if KIC is running in a service mesh:

- namespace where KIC in has certain annotations or labels(e.g.: label 
  `istio-injection=enabled`, annotation `kuma.io/sidecar-injection=enabled`)
  This method is not so accurate, since we can inject to pods using other
  methods, for example, annotating on `deployment` instead of `namespace`.
- annotations or labels in KIC pod or service(e.g.: annotation
 `sidecar.istio.io/status`,`kuma.io/sidecar-injected=true` in KIC pod)
- sidecar container in KIC pod
- initContainer injected by service mesh in KIC pod

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
|signal1: namespace annotation/label| label `istio-injection=enabled` | annotation `linkerd.io/inject=enabled` |annotation `kuma.io/sidecar-injection=enabled`|same as kuma | (NO such annotation/label) | (NO such annotation/label) | `appmesh.k8s.aws/sidecarInjectorWebhook=enabled`|
|signal2: pod/service annotations | pod: `sidecar.istio.io/status` | pod: `linkerd.io/proxy-version` | pod: `kuma.io/sidecar-injected=true` | same as kuma | pod: `consul.hashicorp.com/connect-inject-status=injected` | service: `mesh.traefik.io/traffic-type:(HTTP or TCP)` | NO certain annotations | 
| signal3: sidecar container| `istio-proxy` | `linkerd-proxy`  | `kuma-sidecar` | same as kuma | `envoy-sidecar` | NO sidecar | `envoy` | 
| signal4: init container | `istio-init` | `linkerd-init` | `kuma-init` | same as kuma | `consul-connect-inject-init`  | NO init container | `proxyinit` | 

Special notes:

- kuma/kong mesh can label a `deployment` instead of `namespace` to inject a
  pod.
- kuma/kong mesh does not inject an init container if kuma CNI is deployed.
- traefik does not inject sidecars to pods, but annotate services and use 
  domain `<name>.<namespace>.traefik.mesh` to access a service via mesh.

### Detecting Distribution of Service Mesh

This part proposes methods to detect the numbers of services running in
service mesh. Usually, pods of services running in service mesh are injected
with a sidecar container. We can list all `services` in kubernetes cluster,
and then check their correspoding `endpoints` to get the names of `pods` in
the service. Finally, we check whether the `pods` have a sidecar container to
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

We can assume that service `default/echo` is running in istio mesh, since
there is an `istio-proxy` sidecar container in its pods.

However, traefik mesh is an exception of the method above since it does not
inject sidecar containers. But the method for judging whether a service is 
running in traefik mesh is much simpler. We can check if `service` has one 
of the following annotation: `mesh.traefik.io/traffic-type:HTTP` 
or `mesh.traefik.io/traffic-type:TCP`.

### Format of reporting to splunk

After information of service mesh deployment and distribution are gathered,
we should report the information to our splunk service. Here we propose the
format of messages reported to splunk. Currently, the messages reported to 
splunk is a single line with multiple KV pairs, seperated by `;`. An example
of reported messages is:
```
uptime=120;v=2.3.1;k8sv=1.23.1;db=none;id=1234-5678-abcd-dcba;hn=kong-test
```
Here we integrate all the signals of service meshes into one result and one
message. The key for service mesh deployed is `meshdep`, or `mdep`. The key
for KIC running in service mesh is `kinmesh` or `kinm`. The key for rate of 
services running in mesh is `meshshare` or `mshare` for mesh share. For value
part, we have the following table:

Value used in service mesh deployment part

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
| singnal1: namespace | `istio1` or `i1` | `linkerd1` or `l1` | `kuma1` or `k1` | `kongmesh1` or `km1` | `consul1` or `c1` | `traefik1` or `t1` | `aws1` or `a1` | 
| singnal2: CRD | `istio2` or `i2` | `linkerd2` or `l2` | `kuma2` or `k2` | `kongmesh2` or `km2` | `consul2` or `c2` | None(No unique CRD) | `aws2` or `a2` |
| singnal3: service | `istio3` or `i3` | `linkerd3` or `l3` | `kuma3` or `k3` | `kongmesh3` or `km3` | `consul3` or `c3` | `traefik3` or `t3` | `aws3` or `a3` |

Multiple signals detected would result in multiple parts in comma(`,`)
seperated parts.

For example, if we found `istio-system` and `kuma-system` namespace, CRDs in 
kuma, and `kuma-control-plane` service in `kuma-system`, the key-value pair
would be like: `mdep=i1,k1,k2,k3`.

Values used in KIC running in service mesh part

|Mesh Name| istio  | linkerd   	| kuma   	|   kong mesh	| consul connect | traefik mesh | aws appmesh |
|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:-----:|:------:|
| singnal1: namespace label/annotation | `istio1` or `i1` | `linkerd1` or `l1` | `kuma1` or `k1` | `kongmesh1` or `km1` | `consul1` or `c1` | (NO such annotation) | `aws1` or `a1` | 
| singnal2: pod/service annotation | `istio2` or `i2` | `linkerd2` or `l2` | `kuma2` or `k2` | `kongmesh2` or `km2` | `consul2` or `c2` | `traefik2` or `t2` | `aws2` or `a2` |
| singnal3: sidecar container in pod | `istio3` or `i3` | `linkerd3` or `l3` | `kuma3` or `k3` | `kongmesh3` or `km3` | `consul3` or `c3` | (NO sidecar) | `aws3` or `a3` |
| singnal4: init container in pod | `istio4` or `i4` | `linkerd4` or `l4` | `kuma4` or `k4` | `kongmesh4` or `km4` | `consul4` or `c4` | (NO sidecar) | `aws4` or `a4` |

Multiple signals detected would result in multiple parts in comma(`,`)
seperated parts.

For example, if we found label `istio-injection=enabled` in KIC's namespace,
and `kuma.io/sidecar-injected=true` annotation, `kuma-sidecar` container in
KIC's pod, the message would be `kinm=i1,k2,k3`.

Values in service mesh distribution part

The value would have the following parts: total number of services, mesh
names and number of services running in this mesh. If there are services
running in different service meshes, the numbers are reported seperately. For
example, 50 of 100 services are running in kuma, 25 of 100 services are 
running in istio, we have this message `mshare=100,k50,i25`.

The three parts of detection results are combined to one line seperated by
`;`. For example, the final reported message to splunk would be combined by
the three key-value pairs above. The final message reported would be 
```
mdep=i1,k1,k2,k3;kinm=i1,k2,k3;mshare=100,k50,m25
```
