---
layout: home

hero:
  name: Flagcel
  text: Self-hosted feature flags with CEL targeting.
  tagline: Run the control plane in your own infrastructure, manage flags through a dashboard or API, and evaluate with OpenFeature SDK providers.
  actions:
    - theme: brand
      text: Quickstart
      link: /quickstart
    - theme: alt
      text: View on GitHub
      link: https://github.com/picunada/flagcel

features:
  - title: Self-hosted control plane
    details: Run Flagcel as a Go service backed by Postgres, with migrations embedded into the production binary.
  - title: CEL targeting
    details: Write flag rules with the Common Expression Language used by infrastructure projects such as Kubernetes and Envoy.
  - title: OpenFeature SDKs
    details: Use Go, JavaScript/TypeScript, and Python providers from application code while keeping evaluation keys scoped to environments.
---
