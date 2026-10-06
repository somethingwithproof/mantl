<!-- SPDX-License-Identifier: Apache-2.0 -->
# Mantl examples

For the current compiler, start with [mantl-spec.yaml](mantl-spec.yaml) and
[platform specification/planning](../docs/platform-planning.md). Obtain the matching
release checkout as shown in [the quickstart](../docs/quickstart-platform.md#generate-configuration-offline).
The spec has illustrative identities, a moving Git revision, tenants that must be
committed at tenantPath, and an operator overlay requiring environment setup.
It is an offline generation example, not an unattended deployment recipe.

The application directories contain separate demonstrations; read each local
README and review its requirements before deploying. Their presence does not
establish platform integration or acceptance on every provider.

Older Marathon/Mesos application definitions and Vagrant/Ansible instructions are
historical assets. Use [legacy documentation](../docs/legacy/README.md) only for
that system; they are not a getting-started path for the Go/Kubernetes runtime.
