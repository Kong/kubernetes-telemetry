---
title: Telemetry Library
status: implementable
---

# Telemetry Library

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
- [Alternatives](#alternatives)
<!-- /toc -->

## Summary

Telemetry and analytics are highly important to [Kong][kong] to enable us to
better understand our community and customer use-cases and take action to
improve our products. Traditionally telemetry for products like the [Kong
Kubernetes Ingress Controller (KIC)][kic] have been instrumented in a project
specific manner within the same repository and the data has been shipped to a
long running [Splunk][splunk] server. The purpose of this KEP is to provide
generic tooling for our Kubernetes projects to implement telemetry and to be
more deliberate and transparent in describing the kinds of telemetry data we
collect (and why).

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[splunk]:https://github.com/splunk

## Motivation

- we want to make our telemetry data collection well documented and highly
  transparent for end-users and open source contributors to understand what is
  collected and what value that provides to them
- we've historically had significant gaps in our telemetry for products which
  has made it more difficult to make decisions for those products
- we've never been very deliberate with telemetry data: we don't have clear and
  defined baseline for what we want to collect and what value would be provided
  to end-users who allow us to collect the information.
- previously most of our telemetry code lived in the [KIC][kic], we want to
  reduce the telemetry footprint there and also make telemetry code more generic
  so that it can be used in multiple controllers and other Kubernetes projects.

[kic]:https://github.com/kong/kubernetes-ingress-controller

## Goals

- develop documentation and a standard for our telemetry data
- develop a [Golang][go] library for telemetry

[go]:https://go.dev

## Proposal

To meet the motivations and goals of this KEP we will create a Go library for
configuring and executing workflows that center around access to the Kubernetes
API, processing job results into reports and then shipping those reports to a
(configurable) backend server. For our immediate needs and for historical
familiarity the first backend we will support will be [Splunk][splunk].

The data collected through this library will help us solve a variety of product
and engineering questions. This data helps Kong:

- Monitor and analyze trends, usage, and activities in our Kubernetes products
  (such as the Kong Kubernetes Ingress Controller (KIC))
- Measure the performance of our Kubernetes products
- Research and develop new features based off customer usage

Through this optional library, Kong can better understand our customer
landscape by answering questions such as:

- How many ingress rules are our customers creating?
- Do customers have a service mesh deployed to the cluster, and is the Kong
  Gateway operating inside that network?
- What routing protocols are most widely used?
- Given a certain cluster size or environment, are there any outstanding
  performance issues?

This library will have two main end-user APIs:

1. use directly as a library by other Go projects
2. use via a command line interface (CLI) to enable non-Go projects to run
   telemetry workflows as well (e.g. Helm charts can run workflows as `Jobs` or
   containers in pods).

In the upcoming sections we'll describe how to define `Workflows` from the Go
API and CLI perspective, which will execute the telemetry `Jobs` and send the
reports to Splunk.

[splunk]:https://splunk.com

### Jobs

Under the hood any particular `Job` will be instrumented as a (potentially
repeating) Go function that is started as part of the telemetry framework.

These functions are expressed using the following types:

```go
type Report map[string]interface{}

type Executor func(ctx context.Context, log logrus.Logger, cfg *rest.Config) (Report, error)

type Job struct {
	Name string
	Executor Executor
}
```

Jobs are uniquely identified by name and can be loaded from a catalog of known
jobs at runtime. We will also leave the door open for allowing Go plugins to
extend, though at the time of writing there didn't seem to be any particular
need to support this yet.

### Workflows

Jobs are organized in `Workflows`:

```go
type Workflow struct {
	Jobs        []jobs.Job
	concurrency int
	lock        sync.RWMutex
}

func (wf *Workflow) Run(ctx context.Context, log logrus.Logger, cfg *rest.Config) (jobs.Report, []error) { // ... }
```

Which are threadsafe containers of `Jobs` which can execute jobs concurrently
(while leaving the door open to introduce dependencies later).

### Catalog

When running the telemetry library a number of `Jobs` will be available for
loading via a `Catalog`. The catalog will effectively just be a map of named
jobs to their underlying `Executors` which can be compiled into a `Workflow`.

For the first iteration of this tool, we want to include `Jobs` which can
collect the following information for general environment:

- Orchestration Platform
- Kubernetes Version
- Feature Flags
- Connection to a Service Mesh
- Number of pods
- Number of services
- Architecture

As well as several Kong Specifics:

- Deployment Method
- Last configuration change
- Kong Plugins Count
- Route Types
- Ingress usage (including version)
- Gateway APIs usage
- Knative usage

### Reports

Once a `Workflow` has been compiled from the `Catalog` and comprised of `Jobs`
and the `Workflow` has been run, the result is a `Report` object which will be
shipped to the backend (e.g. Splunk).

Each individual `Job` in a `Workflow` reports it's data as JSON data where the
key for its data is its unique name, e.g.:

```json
{
  "identify-deployment": {
    "lifecycle": "helm",
    "version": "2.8.0",
    "last-update": "2022-05-03 08:46:22"
  },
  "identify-platform": {
    "arch": "amd64",
    "os": "linux",
    "provider": "GKE"
    "kubernetes_version": "v1.23.4",
  }
}
```

### CLI Examples

Running a workflow of several jobs and shipping the report to a specific
Splunk server:

```console
$ ktel workflows run \
    --jobs identify-deployment,identify-platform \
    --report-server splunk://10.0.0.1:8443 \
    --kubeconfig ~/.kube/config
```

The resulting report will be something like:

```json
{
  "identify-deployment": {
    "lifecycle": "helm",
    "version": "2.8.0",
    "last-update": "2022-05-03 08:46:22"
  },
  "identify-platform": {
    "arch": "amd64",
    "os": "linux",
    "provider": "GKE"
    "kubernetes_version": "v1.23.4",
  }
}
```

To support adding additional metrics from a custom source, one can also inject
custom metrics which will be merged into the report, e.g.:

```console
$ echo '{"foo": {"red": "#ff0000", "blue": "#0000ff"}, "bar": {"one": 1}}' | ktel workflows run \
    --jobs identify-deployment,identify-platform \
    --report-server splunk://10.0.0.1:8443 \
    --kubeconfig ~/.kube/config -
```

The result will be the merged report of the workflows run plus the custom data:

```json
{
  "identify-deployment": {
    "lifecycle": "helm",
    "version": "2.8.0",
    "last-update": "2022-05-03 08:46:22"
  },
  "identify-platform": {
    "arch": "amd64",
    "os": "linux",
    "provider": "GKE"
    "kubernetes_version": "v1.23.4",
  }
  "foo": {
    "red": "#ff0000",
    "blue": "#0000ff"
  },
  "bar": {
    "one": 1
  }
}
```

## Alternatives Considered

### Kong Plugins

Originally in this KEP we had expressed a desire to do some in-depth telemetry
on `KongPlugins`, and particularly the use of custom plugins. Doing telemetry
for custom plugins in particular sounds great from a product perspective, but
it also appears to raise some potential security and anonymity concerns. Given
that we very strongly want to maintain anonymity and keep the data we collect
as high level as possible we decided that we would hold off on collecting
information about plugins (beyond just a count) so that we could later
reconsider this and perhas create a KEP just for plugin telemetry and make sure
we define very clear and protective boundaries around what we collect there.

### OpenTelemetry

[OpenTelemetry][ot] has some promise as it helps to standardize telemetry and
also actively works with [Splunk][otsplunk] which is helpful for our needs to
continue using Splunk as our backend for historical reasons, and there's a [Go
SDK][otgo]. The main reason we're so far shy about using OpenTelemetry is its
maturity (for instance, the [metrics API][otmet] is at the time of writing
considered unstable). There's a learning curve and a lot of time that will need
to be spent conforming to OpenTelemetry, and even after that when we do it's
early enough that we may have to spend a considerable amount of time responding
to API changes. For the moment we're going to focus on building something
small, simple and quick but we'll keep OpenTelemetry in the back of our minds
as we continue to iterate.

[ot]:https://opentelemetry.io/docs/instrumentation/go/getting-started/
[otsplunk]:https://www.splunk.com/en_us/blog/conf-splunklive/announcing-native-opentelemetry-support-in-splunk-apm.html
[otgo]:https://github.com/open-telemetry/opentelemetry-go
[otmet]:https://opentelemetry.io/docs/instrumentation/go/manual/#creating-metrics

### Argo Workflows & Jobs

It was considered that we might be able to use [Argo][argo] to implement the
`Workflow` and `Job` components of this project, however our use case was so
narrow that all the extra weight of the full Argo feature set didn't seem to
make sense at the time.

[argo]:https://argoproj.github.io/
