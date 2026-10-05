# ADR 010: Security exception and deployment boundaries

Status: Accepted

## Context

Sonar identified host namespace privileges in the experimental SPIRE agent,
missing load-balancer logging in the unsupported Nomad archive, and public-IP
capabilities in the experimental standalone GCE module. Classifying an issue as
accepted does not establish deployment safety or production readiness.

## Decision

Retain SPIRE's host PID and host network privileges only for the trusted node
agent and its current attestation design. Mandatory verified kubelet TLS,
restricted RBAC and dropped capabilities remain required. Production acceptance
is blocked until an environment-specific review records live attestation,
negative authorization/trust tests, rotation and administrator/socket boundaries,
as listed in the [SPIRE prerequisites](../../infrastructure/cilium/spire-security.md).

The Nomad logging exception applies only to the undeployed, unsupported archive.
Reactivation requires a new supported design with verified logging, TLS and
recorded deployment acceptance. See the
[archive limits](../legacy/nomad/terraform/SECURITY.md).

Remove public-IP capabilities from every GCE VM role and remove the external
legacy service-port firewall. Deprecated inputs reject incompatible public-access
settings. Optional IAP administrative ingress is disabled by default and permits
only logged SSH from the IAP service range to the module's VM tag on its managed
network. Tunnel IAM and SSH authentication require a separate least-privilege
review; the module grants no IAM roles. The module remains experimental and has
no cloud acceptance. Broad internal-subnet rules, external subnet ownership and
outbound connectivity remain explicit deployment review requirements. See the
[GCE module limits](../../infra/terraform/gce/README.md).

## Consequences

The three GCE public-IP findings are addressed in code rather than retained as
risk exceptions. Existing deployments require a reviewed migration and proven
private administration path before applying the removal of public addresses and
firewall rules. No automatic cloud apply or acceptance is authorized by this ADR.

SPIRE's two exceptions and Nomad's archive-only logging exception remain visible
with their conditions, review triggers and evidence requirements. Local checks
verify configuration and failure guards; they do not satisfy live acceptance.
