<!-- SPDX-License-Identifier: Apache-2.0 -->
# ADR 005: Cloud Parity Scope

## Status
Accepted

## Current validation status

The parity decision below is a target for shared static hardening. AWS/GCP/Azure
blueprints have static contract validation, not recorded deployment, identity,
upgrade, audit-delivery, or recovery acceptance. Earlier descriptions of provider
maturity in the context are historical, not current support guarantees. All three
currently use the AWS S3 evidence backend; native GCS/Azure WORM is not implemented.
See [the project maturity table](../../README.md#capabilities-and-maturity) and
[runtime limitations](../architecture-runtime.md#provider-acceptance-and-oscal).

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

## AzureRM 5 identity migration

Azure Key Vault uses RBAC with the AKS encryption identity scoped to its key and
External Secrets scoped to the vault's read-only secrets role. Access-policy
resources are replaced by role assignments; operators must pre-stage, verify and
import the assignments before changing existing vault authorization. Deployment
principals need separately provisioned key-management permission and private-vault
network access. This changes the Azure identity contract without widening cloud
parity claims; backend-disabled validation is the available evidence. See the
Azure blueprint migration procedure for rollout and queue-diagnostics imports.
