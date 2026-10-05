# SPIRE template prerequisites

This template is experimental. Local YAML validation establishes configuration
structure only; it does not establish successful attestation, encrypted application
traffic, certificate rotation or production readiness. Cilium mutual authentication
is an identity check, not application transport encryption. Configure and validate
TLS or an independently supported network encryption layer for sensitive traffic.

Before installing the agent, provision the `spire-kubelet-ca` ConfigMap in `spire`
with `ca.crt` containing the actual kubelet serving-certificate trust roots. There
is deliberately no optional mount, bundled self-signed fallback or verification
bypass. The secure kubelet listener must use port 10250 and its certificate must
validate against the node hostname used by the SPIRE plugin. Review the CA,
certificate SANs, rotation procedure and node DNS identity for the target cluster.
The ordinary Kubernetes service-account token authenticates kubelet requests;
the separate projected `spire-server` audience token is for node attestation.

Enable kubelet token webhook authentication, webhook authorization and
`KubeletFineGrainedAuthz` on every target kubelet. The feature is available before
Kubernetes 1.36 but becomes stable in 1.36; verify it is enabled on the repository's
1.35 target and on managed clusters. Agent RBAC grants `get` on `nodes/pods`, not
`nodes/proxy`. An incompatible kubelet must fail authorization; do not restore
broad proxy access to make attestation succeed. Even `get` on `nodes/proxy` can
permit privileged operations, including container command execution.

The agent retains `hostPID` to resolve workload process/cgroup identity and
`hostNetwork` to reach the local kubelet over loopback. These are agent-specific
privileges, not application defaults. Restrict installation and modification to
trusted platform administrators, isolate the SPIRE namespace, retain the dropped
Linux capabilities and unprivileged container, and review the socket hostPath.
Do not use this template on nodes whose workload or node administrators fall
outside the intended trust boundary. Host namespaces are accepted risks requiring
review whenever the SPIRE architecture or supported node configuration changes.

References: [Cilium mutual authentication limitations](https://docs.cilium.io/en/stable/network/servicemesh/mutual-authentication/mutual-authentication/),
[SPIRE 1.9 workload attestor configuration](https://github.com/spiffe/spire/blob/v1.9.0/doc/plugin_agent_workloadattestor_k8s.md),
[Kubernetes kubelet authentication and authorization](https://kubernetes.io/docs/reference/access-authn-authz/kubelet-authn-authz/).

## Production acceptance remains blocked

The host namespace exceptions are conditional design decisions for a trusted
node identity agent, not production approval. Before deploying this template to
production, record an environment-specific review and live evidence covering:

- Successful node and workload attestation on the supported kubelet configuration,
  including rejection of unauthorized workloads.
- Rejection of an untrusted kubelet CA, invalid server identity and unauthorized
  kubelet access; no verification or RBAC fallback is permitted.
- Certificate renewal and CA rotation without accepting an unverified identity.
- Restricted administrators, namespace modification rights and Workload API socket
  access, with the host namespace and socket trust boundaries reviewed.
- Explicit application TLS or independently validated network encryption for
  sensitive traffic, with failures and recovery recorded.

Record the cluster/version, exact template commit, reviewer and evidence locations.
Local tests do not satisfy these criteria. Review the exceptions again when SPIRE,
Kubernetes, node networking or the administrative trust boundary changes. See
[ADR 010](../../docs/adr/010-security-exception-boundaries.md).
