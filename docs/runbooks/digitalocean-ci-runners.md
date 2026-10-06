<!-- SPDX-License-Identifier: Apache-2.0 -->
# DigitalOcean Ephemeral CI Runners

Mantl CI uses ephemeral DigitalOcean droplets for pushes and manual runs when
both runner credentials are configured. Otherwise it runs the same build and
checks on GitHub-hosted capacity. Pull requests always use GitHub-hosted runners.
There is no standing runner to patch or pay for while idle.

## Safety: CI never touches other droplets

The account runs production workloads. The CI workflows can only ever delete a
droplet that passes all three gates:

1. It carries the dedicated tag `mantl-ci-runner`. Every `doctl` list is filtered
   by this tag, so other droplets are never even enumerated.
2. Its name matches the CI pattern `mantl-ci-<run id>-<run attempt>`. A second check refuses to
   act on any other name.
3. For the reaper, it is older than the orphan threshold (2 hours); a normal run
   lasts minutes.

`stop-runner` deletes only the current run's exact name. The reaper
(`runner-reaper.yml`) only sweeps tagged, CI-named droplets older than the
threshold. Neither ever deletes by anything broad.

## Teardown is guaranteed three ways

1. The runner registers with `--ephemeral`, so it deregisters itself from GitHub
   after one job.
2. `stop-runner` runs with `if: always()`, so the droplet is destroyed even when
   the build fails or the run is cancelled. It deletes by deterministic name, not
   by a value passed from an earlier job, so it works even if `start-runner`
   failed partway. Including the run attempt keeps reruns isolated from stale
   droplets created by an earlier attempt.
3. The scheduled reaper deletes any orphan that slipped past the first two (for
   example if the `stop-runner` runner itself was cancelled).

## Optional DigitalOcean runner secrets

| Secret | What it is | Scope |
|--------|-----------|-------|
| `DIGITALOCEAN_ACCESS_TOKEN` | DO API token for `doctl` | Write (create/delete droplets) |
| `GH_RUNNER_PAT` | GitHub fine-grained PAT to mint runner registration tokens | Repository `Administration: read and write` on this repo only |

The default `GITHUB_TOKEN` cannot create runner registration tokens, which is why
a scoped PAT is needed. Keep the PAT scoped to this repository.

Add both under Settings to Secrets and variables to Actions.

The reaper checks for `DIGITALOCEAN_ACCESS_TOKEN` before installing `doctl` or
contacting DigitalOcean. If it is absent, the run succeeds with a notice and a
job summary stating that cleanup was skipped; no droplets are listed or
deleted. This supports forks that have not enabled DigitalOcean runners.
Provision the token before relying on scheduled cleanup. A configured but
invalid token still fails authentication normally.

## Droplet sizing

Set in `ci.yml` env: `CI_RUNNER_SIZE` (default `s-4vcpu-8gb`), `CI_RUNNER_IMAGE`,
`CI_RUNNER_REGION`. 4 vCPU / 8 GB gives `golangci-lint` and the Go build (k8s,
aws-sdk, controller-runtime, prometheus dependencies) enough headroom; 4 GB risks
an OOM. Because droplets are destroyed in minutes, a larger size costs only
pennies per run.

## Making CI a required check

Missing either secret selects the GitHub-hosted backend before provisioning.
The job summary records the selected backend without revealing credentials.
Configured but invalid credentials fail the cloud runner path rather than
silently falling back. The build must succeed on either backend for `CI` to pass.

After one full run goes green, add the `CI` check to
`main` branch protection so it becomes a required gate, matching the sibling
repo's posture:

```bash
gh api -X PATCH \
  repos/somethingwithproof/mantl/branches/main/protection/required_status_checks \
  -f 'contexts[]=DCO' -f 'contexts[]=CI'
```

The `CI` check is the aggregate `ci` job, which passes only when the `build` job
on the selected runner passes; configured cloud runs must also pass provisioning
and cleanup. Cleanup runs even after provisioning partially fails, using the
same exact-name and dedicated-tag filters described above.
