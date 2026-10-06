<!-- SPDX-License-Identifier: Apache-2.0 -->
# DigitalOcean runners

Eligible Linux jobs select single-job DigitalOcean runners when the repository
variable `DO_RUNNERS_ENABLED` is `true`. An unset or false value selects
GitHub-hosted runners. The dedicated labels are `self-hosted`, `Linux`, `X64`
and `mantl-do`; use repository-scoped registrations, not another project's pool.

Push, schedule, manual dispatch and same-repository pull requests are eligible.
Fork pull requests, privileged PR metadata events, comment events and standalone
release events retain GitHub-hosted routing. The external SLSA generator controls
its own runners. The legacy orphan reaper stays GitHub-hosted because its cloud
credential must not enter a workload runner.

The dedicated Mantl controller uses the existing GitHub App installation for
somethingwithproof, but restricts its token and public-repository allowlists to
`somethingwithproof/mantl`. The installation router matches both installation
identity and repository before forwarding the original signed delivery.
The controller verifies webhook signatures and GitHub's stored run repository,
numeric identities and non-fork origin before allocating capacity. Missing or
conflicting identities and unavailable verification fail closed.

Each fresh VM receives a single-use JIT configuration and executes one job.
Cloud credentials and the App key remain on the controller. Runner Docker and
noninteractive sudo access are root-equivalent inside the disposable VM.
Use a dedicated VPC and a tag-scoped firewall with no inbound rules; permit
outbound HTTP/HTTPS and DNS for signed package repositories and GitHub services.
Pin and verify runner downloads, provide Docker/buildx, gh and a writable
`/opt/hostedtoolcache`, and select all job runtimes through mise.

Run **DigitalOcean runner smoke** before cutover. Its same-repository PR trigger
allows provisioning validation while the routing variable remains false;
forks cannot run the smoke job. Verify runner prerequisites, mise runtime setup,
Docker execution, completion and actual VM deletion. Power-off alone does not
end DigitalOcean billing. A job timeout starts after runner assignment: cancel
queued smoke runs if provisioning fails.

Manual smoke runs accept `jobs=1` (default) or `jobs=8`. Use the eight-job mode
to verify queued demand beyond the normal four-runner cap:

```sh
gh workflow run digitalocean-runner-smoke.yml \
  --repo somethingwithproof/mantl --ref main --field jobs=8
```

All eight samples must complete, and every allocated VM must be confirmed gone.
A JIT runner can execute a different queued job than its trigger. The controller
must preserve the original demand when that happens, replace capacity only after
cleanup and a fresh queued-job check, and retain authoritative original-job
completion across restart. PR-triggered smoke runs use one sample.

After smoke and deletion succeed, set `DO_RUNNERS_ENABLED=true` and run ordinary
CI against the current PR head. Enable release routing only with successful
package/image validation; exact-tag checks, signing and permissions remain
required. No cloud acceptance for Mantl clusters is established by these runners.

Rollback by setting `DO_RUNNERS_ENABLED=false`, cancelling queued self-hosted
runs and dispatching new runs. Already queued jobs retain their original runner
selection, and historical workflow revisions do not acquire new routing rules.
Drain active runners before stopping the dedicated controller. Its state and
ownership tags provide cleanup independently of the legacy orphan reaper.
