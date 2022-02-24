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
Kubernetes Ingress Controller (KIC)][kic] have been instrumented in a project
specific manner within the same repository and the data has been shipped to a
long running [Splunk][splunk] server. The purpose of this KEP is to be more
deliberate in describing the kinds of telemetry data we want from our
Kubernetes products, and providing generic tooling to handle this between all
of our projects.

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
- previously most of our telemetry code lived in the [KIC][kic], we want to
  reduce the telemetry footprint there and also make telemetry code more generic
  so that it can be used in multiple controllers.

[kic]:https://github.com/kong/kubernetes-ingress-controller

### Goals

- develop documentation and a standard for our telemetry data
- develop a [Golang][go] library for telemetry

[go]:https://go.dev

## TODO

(sean): we need some process and examples which would help us show end-users
_how_ we will end up using this data and what effects it will have on their
products which will benefit them.
