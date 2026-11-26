<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lisensya: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Nauulit na Kubernetes platform configuration na may masusubaybayang ebidensya ng pagsunod.

## Pangkalahatang-ideya

Ang Mantl ay isang spec-driven na platform compiler at compliance runtime. Ang isang `MantlCluster` YAML specification ay kino-compile sa mga Terraform variable input, ArgoCD Application manifest, at tenant namespace isolation manifest. Isang hiwalay na Go compliance operator ang nagre-reconcile ng mga control mapping, finding, evaluation, at naka-bersyong S3 audit evidence.

- Tumatakbo nang offline ang `plan` upang mag-compile, mag-save, o maghambing ng mga platform artifact nang hindi binabago ang imprastraktura.
- Nagsasagawa ng mga preflight check ang `apply` at nagbu-bootstrap ng mga nabuo o nasuring platform artifact.
- Nagmamasid ang `status` ng mga nabuong application at bilang ng mga Kyverno policy report.
- Sinusuri ng `compliance` ang compliance coverage, mga finding, remediation, history, at mga archive ng ebidensya.

## Mabilisang pagsisimula

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Halimbawang `platform.yaml`:

```yaml
apiVersion: platform.mantl.io/v1alpha1
kind: MantlCluster
metadata:
  name: team-dev
spec:
  provider:
    kind: aws
    region: us-east-1
  kubernetes:
    distribution: eks
    version: "1.31"
  profile:
    size: small
```

## Lokalisasyon

Nag-e-embed ang `mantl` CLI ng mga salin para sa 45 locale (20 Tier 1, 15 African, at 10 Asian na locale) sa `pkg/i18n`. Sinusunod ng pagpili ng locale ang `LC_ALL`, `LC_MESSAGES`, `LANG`, at ang persistent na `--lang` flag:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Pag-develop at Lisensya

Patakbuhin ang mga Go unit test at check sa pamamagitan ng `make go-test`, `make go-coverage`, at `golangci-lint run`. Tingnan ang buong [README sa Ingles](../../../README.md) at [indeks ng dokumentasyon](../../../docs/README.md). Lisensyado sa ilalim ng [Apache-2.0](../../../LICENSE).
