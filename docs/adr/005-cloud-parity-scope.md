# ADR 005: Cloud Parity Scope

## Status
Accepted

## Context
The README advertised seven clouds. Only AWS EKS was production-grade; GCP and
Azure were bare module wrappers, and the rest were unverified. Maintaining seven
hardened blueprints is not realistic for the current team.

## Decision
Parity targets three clouds: AWS (EKS), GCP (GKE), and Azure (AKS), brought to a
common hardening baseline (private cluster, control-plane audit logging,
encryption at rest, workload identity, least-privilege IAM). Linode, Oracle, IBM,
and DigitalOcean are labeled experimental in docs and directory layout.

## Rationale
Three credible clouds serve more real users than seven half-built ones. Honest
maturity labels protect the project's credibility.

## Consequences
Phase 3 hardens GCP and Azure to the AWS baseline. Experimental blueprints carry
no parity or CI guarantees until a contributor adopts them.
