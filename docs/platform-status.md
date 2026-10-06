<!-- SPDX-License-Identifier: Apache-2.0 -->
# Platform status (beta)

`mantl status` observes the ArgoCD applications generated from a supplied spec and
summarizes policy engine results. It performs read-only Kubernetes queries; it does
not bootstrap, reconcile, or modify resources.

```sh
mantl status examples/mantl-spec.yaml --context staging
mantl status examples/mantl-spec.yaml --context staging --format json
mantl status examples/mantl-spec.yaml --context staging --format json --require-healthy
```

The last command prints the same report and returns a nonzero exit status unless
every expected generated application is `Synced` and `Healthy`. Query failures,
missing applications, duplicate observations, incomplete status, and mismatched
source or destination identities cannot satisfy this gate. ArgoCD must have
compared the application against the configured repository, revision, path and
destination. Error conditions and unfinished or failed operations also prevent a
healthy result. Unrelated applications neither substitute for required applications
nor change their result.

The spec and renderer share the desired application topology: `root-platform`,
enabled feature applications, and `platform-tenants` when tenants or a tenant source
path are configured. The gate covers these directly generated applications. It does
not recursively check the root application's children, tenant workloads, rollout
availability, cloud infrastructure, or recovery. An ArgoCD observation is a snapshot,
not independent proof that a moving Git branch points at the latest desired commit.
A successful gate does not establish production readiness.

Both queries share a five-second deadline; `--timeout` accepts a positive duration
up to one minute. Cancellation propagates from the command. Responses are limited
to 8 MiB per query, and inherited process pipes have a one-second drain limit. Raw
application specs and external command errors are excluded from output, including
credential-plugin stderr. A query failure is reported as `Unknown`; normal
observation mode still succeeds so callers can inspect the result. Invalid input,
cancellation, output errors, and a failed explicit health gate return an error.

## JSON contract

The schema version is `mantl.io/status/v1alpha1`. The report contains `observedAt`
(UTC report completion time), `platform` (spec name), `context` (requested Kubernetes
context), `applications`, and `policies`. An empty context means kubectl uses the
current context; it does not assert which cluster was selected. Consumers should
check the schema version and application `state`, rather than parse text output.

Application states are `Healthy`, `Degraded`, or `Unknown`. `Unknown` takes priority
when desired applications cannot be evaluated completely. Each evaluated application
has a name, state, and optional sync, health, and reason fields. Query or list parsing
failures can leave the application list empty; the aggregate state remains `Unknown`.

Policy scope is `all-readable-policy-reports`: namespaced PolicyReports and
ClusterPolicyReports available through the selected credentials, across namespaces.
These results are not filtered to a Mantl tenant or application. Missing permissions
or CRDs produce `Unknown`. With data, counts distinguish `pass`, `fail`, `error`,
`warn`, and `skip`; there is no percentage or regulatory compliance claim. Missing,
negative, fractional, or overflowing summary counts are rejected. Reports without
results remain `Unknown`, and absent or invalid data yields `counts: null`.

The application gate does not assess policy results. Use Mantl's audit/evaluation
contracts to assess framework coverage, evidence freshness and approved exceptions;
PolicyReport summaries cannot replace those contracts.
