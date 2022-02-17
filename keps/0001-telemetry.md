---
title: Telemetry Library
status: provisional
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
Kubernetes Ingress Controller (KIC)][kic] has been instrumented in a project
specific manner within the same repository and the data has been shipped to a
long running [Splunk][splunk] server. The purpose of this KEP is to be more
deliberate in describing the kinds of telemetry data we want from our
Kubernetes products, and providing generic tooling to handle this between all
of our projects.

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[splunk]:https://github.com/splunk

## Motivation

- we've historically had significant gaps in our telemetry for products which
  has made it more difficult to make decisions for those products
- we've never been very deliberate with telemetry data: we don't have clear and
  defined baseline for what we want to collect and what value would be provided
- we want to make our telemetry data collection well documented and highly
  transparent for end-users and open source contributors to understand what is
  collected and what value that provides to them
- previously most of our telemetry code lived in the [KIC][kic], we want
  standard tooling that can be used by multiple projects and provides a wide
  range of common functionality
- we want testing tools that allow us to bring up a test version of our
  telemetry infrastructure so that downstream projects can easily write
  integration tests to validate their telemetry functionality

[kic]:https://github.com/kong/kubernetes-ingress-controller

### Goals

- develop documentation and a standard definining the telemetry data we want to
  be able to ideally provide for each project.
- develop a [Golang][go] library which provides an API for the standard that
  can be imported and used in other projects, and includes facilities for
  integration testing.

[go]:https://go.dev
