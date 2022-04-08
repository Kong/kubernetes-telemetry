
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

## Telemetry Collection

The data collected through this library will help us solve a variety of product and engineering questions. This data helps Kong:

 - Monitor and analyze trends, usage, and activities while using KIC
 - Measure the performance of the KIC
 - Research and develop new features based off customer usage

Through this optional library, Kong can better understand our customer landscape by answering questions such as:

 - How many ingress rules are our customers creating?
- Do customers have a service mesh deployed no the cluster, and is the Kong Gateway operating inside that network?
 - What routing protocols are most widely used?
 - Given a certain cluster size or environment, are there any outstanding performance issues?

More specifically this tool will have the ability to collect the following pieces of information:
 
**General Environment**
 - Orchestration Platform
 - Kubernetes Version
 - Feature Flags
 - Connection to a Service Mesh
 - Number of clusters
 - Number of pods
 - Number of services
 - Architecture
 
 **Kong Specifics**
 
 - Kong Plugins
	 - Count
	 - Custom Plugins
 - Deployment Method
 - Route Types
 - Ingress version
 - Last configuration change
 - Gateways

 ## Acceptance Criteria

 - [ ] Golang library is published
 - [ ] All telemetry outlined in Telemetry Collection is collected and sent to splunk
 - [ ] An opt out setting is created

 
