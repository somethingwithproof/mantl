<!-- SPDX-License-Identifier: Apache-2.0 -->
# Backstage integration

Register `templates/backstage/web-service/template.yaml` in an existing Backstage
catalog. Enable the `fetch:template` and `publish:github:pull-request` scaffolder
actions with a scoped GitHub integration. The template renders its local skeleton
and opens a pull request at `apps/services/<name>` in the selected GitOps repository.

Choose an existing tenant namespace and a verified ghcr.io image digest. Generated
workloads expose `/healthz` on port 8080, run as non-root, drop capabilities, and
have resource limits. Review the pull request, provide explicit tenant network
policies, then register its catalog-info.yaml in Backstage. No live portal service
or automatic application deployment is enabled by this integration.
