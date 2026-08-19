# DigitalOcean Ephemeral CI Runners

Mantl CI runs on ephemeral DigitalOcean droplets: one is created per CI run,
runs the build, and is destroyed when the run finishes. There is no standing
runner to patch or pay for while idle.

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

## Required secrets (provision before CI can run)

| Secret | What it is | Scope |
|--------|-----------|-------|
| `DIGITALOCEAN_ACCESS_TOKEN` | DO API token for `doctl` | Write (create/delete droplets) |
| `GH_RUNNER_PAT` | GitHub fine-grained PAT to mint runner registration tokens | Repository `Administration: read and write` on this repo only |

The default `GITHUB_TOKEN` cannot create runner registration tokens, which is why
a scoped PAT is needed. Keep the PAT scoped to this repository.

Add both under Settings to Secrets and variables to Actions.

## Droplet sizing

Set in `ci.yml` env: `CI_RUNNER_SIZE` (default `s-4vcpu-8gb`), `CI_RUNNER_IMAGE`,
`CI_RUNNER_REGION`. 4 vCPU / 8 GB gives `golangci-lint` and the Go build (k8s,
aws-sdk, controller-runtime, prometheus dependencies) enough headroom; 4 GB risks
an OOM. Because droplets are destroyed in minutes, a larger size costs only
pennies per run.

## Making CI a required check

Until the secrets above exist, CI runs will fail at `start-runner` (no token).
That is expected and does not block merges, because branch protection requires
DCO and conversation resolution, not CI yet.

After the secrets are set and one run goes green, add the `CI` check to
`main` branch protection so it becomes a required gate, matching the sibling
repo's posture:

```bash
gh api -X PATCH \
  repos/somethingwithproof/mantl/branches/main/protection/required_status_checks \
  -f 'contexts[]=DCO' -f 'contexts[]=CI'
```

The `CI` check is the aggregate `ci` job, which passes only when the `build` job
on the DO runner passes.
