<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Igenamiterere ry'urubuga rwa Kubernetes rishobora gusubirwamo rifite ibimenyetso by'iyubahirizwa ry'amabwiriza bikurikiranwa.

## Incamake

Mantl ni umuhuzabikorwa w'urubuga ushingiye ku bisobanuro n'uburyo bwo kugenzura iyubahirizwa ry'amabwiriza. Ibisobanuro bya YAML bya `MantlCluster` bihindurwa mo ibyinjizwa by'impinduragaciro za Terraform, inyandiko za ArgoCD Application, n'inyandiko zo gutandukanya amazina y'abakoresha. Umukozi wa Go ushinzwe iyubahirizwa ry'amabwiriza utandukanye ahuza amakarita y'ubugenzuzi, ibyavumbuwe, isuzuma, n'ibimenyetso by'igenzura bya S3 bifite verisiyo.

- `plan` ikora idafite internet kugira ngo itunganye, ibike, cyangwa igereranye ibikoresho by'urubuga idahinduye ibikorwaremezo.
- `apply` ikora igenzura ry'ibanze kandi igatangiza ibikoresho by'urubuga byakozwe cyangwa byasuzumwe.
- `status` igenzura porogaramu zakozwe n'imibare ya raporo za politiki ya Kyverno.
- `compliance` isuzuma urugero rw'iyubahirizwa ry'amabwiriza, ibyavumbuwe, ikosora, amateka, n'ububiko bw'ibimenyetso.

## Gutangira vuba

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Urugero rwa `platform.yaml`:

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

## Ishyirwa mu ndimi

CLI ya `mantl` irimo ubusobanuro bw'indimi 45 (indimi 20 z'icyiciro cya 1, 15 zo muri Afurika, na 10 zo muri Aziya) muri `pkg/i18n`. Ihitamo ry'ururimi rikurikiza `LC_ALL`, `LC_MESSAGES`, `LANG`, n'ikimenyetso gihoraho cya `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Iterambere n'Uruhushya

Koresha ibizamini by'ibice bya Go n'igenzura ukoresheje `make go-test`, `make go-coverage`, na `golangci-lint run`. Reba [README yuzuye mu Cyongereza](../../../README.md) n'[urutonde rw'inyandiko](../../../docs/README.md). Ifite uruhushya rwa [Apache-2.0](../../../LICENSE).
