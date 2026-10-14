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

## Wait for application health

Current source adds a read-only deployment gate that polls until every generated
application satisfies the same strict identity, sync and health checks:

```sh
mantl --context staging status wait examples/mantl-spec.yaml
mantl --context staging status wait examples/mantl-spec.yaml \
  --timeout 15m --interval 5s --query-timeout 5s --format json
```

This subcommand is not included in the published v0.5.0 CLI. It does not apply
resources or change the current kube-context. The default overall deadline is
ten minutes; `--timeout` accepts a positive duration up to two hours. Each pair of
queries has a separate `--query-timeout` (five seconds by default, at most one
minute), bounded by the remaining overall deadline. `--interval` pauses between
observations, defaults to five seconds and accepts one second through one minute.
Unlike one-shot status, the wait command's `--timeout` governs the entire poll.

Transient query failures, missing applications, unknown status and degraded
applications are retried until the deadline. Missing or mismatched identities
cannot satisfy the gate merely because some other application is healthy. A
successful exit requires every expected application to be synced and healthy.
Policy counts are included for observation but do not determine wait success;
missing policy permissions or failed policy results cannot become a compliance
verdict. The scope remains the directly generated applications described above.

The command writes one final report in the existing status schema, without
streaming intermediate JSON objects. On timeout or cancellation after an
observation, it prints the last observation and returns a nonzero exit status.
Invalid input or cancellation before any query produces no observation report.
Check the command exit status as well as the application state; a timeout or
cancellation cannot establish success. External errors remain sanitized, and
each response retains the same size and process-pipe bounds as one-shot status.

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
