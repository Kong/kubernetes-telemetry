# Kong Kubernetes Telemetry

Toolkit for telemetry data for [Kong's][kong] Kubernetes products, such as the
[Kong Kubernetes Ingress Controller (KIC)][kic].

## Usage

```go
import (
  "context"
  "time"

  "github.com/bombsimon/logrusr/v3"
  "github.com/sirupsen/logrus"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"

  "github.com/kong/kubernetes-telemetry/pkg/forwarders"
  "github.com/kong/kubernetes-telemetry/pkg/serializers"
  "github.com/kong/kubernetes-telemetry/pkg/telemetry"
)

func main() {
  log := logrusr.New(logrus.New())
  m, err := telemetry.NewManager(
    "custom-ping",
    telemetry.OptManagerPeriod(time.Hour),
    telemetry.OptManagerLogger(log),
  )
  // Handle errors ...

  // Configure your Kubernetes client(s)
  loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
  // If you want to change the loading rules (which files in which order), you can do so here
  kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, nil)
  restConfig, err := kubeConfig.ClientConfig()
  // Handle errors ...
  cl, err := kubernetes.NewForConfig(restConfig)
  // Handle errors ...

  w, err := telemetry.NewIdentifyPlatformWorkflow(cl)
  // Handle errors ...
  m.AddWorkflow(w)

  // Add more workflows/providers if needed ...
  
  // Configure serialization ...
  serializer := serializers.NewSemicolonDelimited()
  // ... and forwarding
  tf, err := forwarders.NewTLSForwarder(splunkEndpoint, log)
  // Handle errors ...
  consumer := telemetry.NewConsumer(serializer, tf)
  m.AddConsumer(consumer)

  // Start the manager
  err := m.Start()
  // Handle errors ...


  // Trigger asynchronous report as needed.
  err := m.TriggerExecute(context.Background(), "custom-event-happened");
  // Handle errors ...
```

### Forwarders

Forwarders can be used to forward serialized telemetry reports to a particular destination.

- `TLSForwarder` can be used to forward data to a TLS endpoint
- `LogForwarder` can be used to forward data to a configured logger instance
- `DiscardForwarder` can be used to discard received reports

### Serializers

Users can pick the serializer of their choice for data serialization.

Currently only 1 serializer is supported with more implementations to come as needed.

#### Semicolon delimited values

This serializer uses the following predefined keys to express telemetry data:

- `k8s_arch` - inferred kubernetes cluster architecture
- `k8sv` - inferred kubernetes cluster version
- `k8sv_semver` - inferred kubernetes cluster version in [semver format][semver]
- `k8s_provider` - inferred kubernetes cluster provider
- `k8s_pods_count` - number of pods running in the cluster
- `k8s_services_count` - number of services defined in the cluster
- `hn` - hostname where this telemetry framework is running on
- `feature-<NAME>` - feature gate (with the boolean state indicated whether enabled or disabled)

[kong]:https://github.com/kong
[kic]:https://github.com/kong/kubernetes-ingress-controller
[semver]:https://semver.org/
