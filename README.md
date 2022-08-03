# Kong Kubernetes Telemetry

Toolkit for telemetry data for [Kong's][kong] Kubernetes products, such as the
[Kong Kubernetes Ingress Controller (KIC)][kic].

## Usage

### Serializers

Users can pick the serializer of their choice for data serialization.

Currently only 1 serializer is supported with more implementations to come as needed.

#### Semicolon delimited values

This serializer uses the following predefined keys to express telemetry data:

* `k8sarch` - inferred kubernetes cluster architecture
* `k8sv` - inferred kubernetes cluster version
* `k8sv_semver` - inferred kubernetes cluster version in [semver format][semver]
* `k8s_provider` - inferred kubernetes cluster provider
* `k8s_pods` - number of pods running in the cluster
* `k8s_services` - number of services defined in the cluster
* `hn` - hostname where this telemetry framework is running on
* `feature-<NAME>` - feature gate (with the boolean state indicated whether enabled or disabled)

[kong]:https://github.com/kong
[kic]:https://github.com/kong/kubernetes-ingress-controller
[semver]:https://semver.org/
