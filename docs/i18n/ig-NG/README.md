<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Nhazi ikpo okwu Kubernetes a na-emegharị emegharị nke nwere ihe akaebe nnabata iwu a pụrụ ịchọta.

## Nchịkọta

Mantl bụ onye nchịkọta ikpo okwu dabere na nkọwapụta yana oge mmezu nnabata iwu. Nkọwapụta YAML `MantlCluster` na-achịkọta n'ime ntinye mgbanwe Terraform, akwụkwọ ngwa ArgoCD Application, na akwụkwọ nkewa oghere aha ndị bi. Onye ọrụ nnabata iwu Go dị iche na-edozi njikọ njikwa, nchọpụta, nyocha, na ihe akaebe nyocha S3 nwere ụdị dị iche iche.

- `plan` na-agba n'enweghị njikọ ịntanetị iji chịkọta, chekwaa, ma ọ bụ tụnyere arụrụ ikpo okwu n'agbanweghị akụrụngwa.
- `apply` na-agba nyocha mbido ma na-ebido arụrụ ikpo okwu e mepụtara ma ọ bụ tụlere.
- `status` na-eleba anya na ngwa e mepụtara na ọnụọgụ mkpesa iwu Kyverno.
- `compliance` na-enyocha mkpuchi nnabata iwu, nchọpụta, mmezi, akụkọ ihe mere eme, na ebe nchekwa ihe akaebe.

## Mmalite ngwa ngwa

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Ọmụmaatụ `platform.yaml`:

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

## Nkwado asụsụ obodo

CLI `mantl` tinyere ntụgharị asụsụ maka asụsụ obodo 45 (asụsụ Tier 1 20, asụsụ Afrịka 15, na asụsụ Eshia 10) na `pkg/i18n`. Nhọrọ asụsụ obodo na-asọpụrụ `LC_ALL`, `LC_MESSAGES`, na `LANG`, yana ọkọlọtọ `--lang` na-adịgide adịgide:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Mmepe na Ikikere

Gbaa ule nkeji Go na nyocha site na `make go-test`, `make go-coverage`, na `golangci-lint run`. Lee [README Bekee](../../../README.md) zuru ezu na [ndepụta akwụkwọ ntụziaka](../../../docs/README.md). E nyere ikikere n'okpuru [Apache-2.0](../../../LICENSE).
