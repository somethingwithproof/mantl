<!-- SPDX-License-Identifier: Apache-2.0 -->
# License declarations

Mantl's first-party code and documentation use `SPDX-License-Identifier: Apache-2.0`,
matching [LICENSE](../LICENSE), except the existing MIT components below. Existing copyright notices stay intact. The static
website example and WordPress scaffold retain their existing MIT declarations; the unmodified Developer
Certificate of Origin uses `LicenseRef-DCO`.

Use the language's comment syntax near the start of source files. Keep executable
shebangs, Docker parser directives, XML declarations and Go build constraints in
their required positions. Markdown/HTML/SVG use hidden comments. JSON, binaries,
public keys, test fixtures and generated CRD data use adjacent `.license` files
so their bytes and parsers remain unchanged. These declarations follow
[SPDX short-form guidance](https://spdx.dev/learn/handling-license-info/) and
[REUSE companion-file/annotation conventions](https://reuse.software/spec-3.3/).

`make generate` supplies the SPDX boilerplate to controller-gen. Python dependency
lock regeneration emits the same header as the checked-in locks. The legacy
DeepCopy file uses a companion declaration rather than a manual generated-code edit.
Content headers change bundle hashes, so this update uses control-content version
`1.0.1`; published `1.0.0` content must remain immutable.

## Vendored content

[REUSE.toml](../REUSE.toml) annotates versioned chart trees without editing upstream
bytes. Its `closest` precedence preserves any declaration inside an upstream file.
Add a reviewed annotation when introducing a new chart/version; there is no blanket
fallback for an unknown chart path.

| Chart source | Declared chart-file license | Evidence |
| --- | --- | --- |
| Cilium, Crossplane, Harbor | Apache-2.0 | Included upstream LICENSE files |
| cert-manager | Apache-2.0 | [v1.16.2 license](https://github.com/cert-manager/cert-manager/blob/v1.16.2/LICENSE) and chart annotation |
| external-dns | Apache-2.0 | [v0.15.0 license](https://github.com/kubernetes-sigs/external-dns/blob/v0.15.0/LICENSE.md) |
| OpenTelemetry collector | Apache-2.0 | [0.107.0 chart-source license](https://github.com/open-telemetry/opentelemetry-helm-charts/blob/opentelemetry-collector-0.107.0/LICENSE) |
| Prometheus stack and its exporter/Grafana charts | Apache-2.0 | [65.2.0 chart-source license](https://github.com/prometheus-community/helm-charts/blob/kube-prometheus-stack-65.2.0/LICENSE) and embedded chart annotations |
| Loki | Apache-2.0 | [`production/` exception at v3.2.0](https://github.com/grafana/loki/blob/v3.2.0/LICENSING.md) |
| Grafana agent operator, rollout operator, Tempo | Apache-2.0 | [agent 0.5.0](https://github.com/grafana/helm-charts/blob/grafana-agent-operator-0.5.0/LICENSE), [rollout 0.20.0](https://github.com/grafana/helm-charts/blob/rollout-operator-0.20.0/LICENSE), [Tempo 1.14.0](https://github.com/grafana/helm-charts/blob/tempo-1.14.0/LICENSE) |
| MinIO dependency inside Loki | AGPL-3.0-or-later | [pinned source license](https://github.com/minio/minio/blob/RELEASE.2024-04-18T19-09-19Z/LICENSE) and [or-later notice](https://github.com/minio/minio/blob/RELEASE.2024-04-18T19-09-19Z/main.go) |
| External Secrets and Bitwarden SDK server chart | Apache-2.0 | [v0.12.1](https://github.com/external-secrets/external-secrets/blob/v0.12.1/LICENSE), [SDK server v0.3.1](https://github.com/external-secrets/bitwarden-sdk-server/blob/v0.3.1/LICENSE) |
| Falco and bundled charts | Apache-2.0 | [Falco chart 4.15.0 source license](https://github.com/falcosecurity/charts/blob/falco-4.15.0/LICENSE) and existing file headers |
| Kyverno and bundled wrappers | Apache-2.0 | [3.3.3 chart-source license](https://github.com/kyverno/kyverno/blob/kyverno-chart-3.3.3/LICENSE) |

These declarations describe the files in this checkout. Referenced container images
and deployed software retain their own licenses; a chart license does not assign
the license of its runtime image.

## Check and maintain

```sh
mise exec -- python -m ci.check_license_headers
```

The check covers tracked files, companion files and the supported REUSE annotation
subset (exact paths and `directory/**`, with `closest` precedence). Each declared
identifier needs its corresponding text in `LICENSES/`. Existing upstream notices
take precedence. The check verifies license declarations and accompanying texts;
copyright attribution remains a separate review responsibility.

Keep companion declarations with their files when moving or deleting them. Preserve
license texts verbatim. Run the normal language/configuration checks after header
changes, and regenerate generated files through their owning tools.
