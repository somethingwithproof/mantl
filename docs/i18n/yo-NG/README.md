<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Ìwé-àṣẹ: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Ìṣètò pẹpẹ Kubernetes tí a lè tún ṣe pẹ̀lú ẹ̀rí ìbámu-ìlànà tí a lè tọ́pinpin.

## Àkópọ̀

Mantl jẹ́ òǹṣàkójọpọ̀ pẹpẹ tí ó dá lórí ìwé-àpèjúwe àti àkókò-ìṣiṣẹ́ ìbámu-ìlànà. Ìwé-àpèjúwe YAML `MantlCluster` máa ń ṣàkójọpọ̀ di àwọn ohun-ìtẹ̀síwájú oníyípadà Terraform, àwọn ìwé-àfihàn Ohun-èlò ArgoCD, àti àwọn ìwé-àfihàn ìyàsọ́tọ̀ ààyè-orúkọ olùlò. Olùṣàkóso ìbámu-ìlànà Go ọ̀tọ̀ kan ń mú ìbáramu àwọn ìṣàkóso, àwọn àwárí, àwọn ìgbéléwọ̀n, àti ẹ̀rí àyẹ̀wò S3 tí ó ní ẹ̀dà ṣẹ.

- `plan` ń ṣiṣẹ́ láìsí ìsopọ̀ ayélujára láti ṣàkójọpọ̀, tọ́jú, tàbí fi àwọn ohun-ìṣẹ̀dá pẹpẹ wéra láìyí àwọn ohun-ìpìlẹ̀ padà.
- `apply` ń ṣiṣẹ́ àwọn àyẹ̀wò ìṣáájú, ó sì ń bẹ̀rẹ̀ àwọn ohun-ìṣẹ̀dá pẹpẹ tí a ṣẹ̀dá tàbí tí a ṣàyẹ̀wò.
- `status` ń ṣe àkíyèsí àwọn ohun-èlò tí a ṣẹ̀dá àti iye ìròyìn ìlànà Kyverno.
- `compliance` ń ṣe àyẹ̀wò ààbò ìbámu-ìlànà, àwọn àwárí, àtúnṣe, ìtàn, àti àwọn àkájọ ẹ̀rí.

## Ìbẹ̀rẹ̀ kíákíá

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Àpẹẹrẹ `platform.yaml`:

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

## Ìsọdèdè-agbègbè

CLI `mantl` fi àwọn ìtumọ̀ kún inú rẹ̀ fún àwọn èdè agbègbè 45 (20 ti Ìpele 1, 15 ti Áfíríkà, àti 10 ti Éṣíà) nínú `pkg/i18n`. Yíyan èdè agbègbè ń tẹ̀lé `LC_ALL`, `LC_MESSAGES`, `LANG`, àti àsíá `--lang` tí ó ń dúró:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Ìdàgbàsókè àti Ìwé-àṣẹ

Ṣiṣẹ́ àwọn ìdánwò ẹ̀yọ Go àti àwọn àyẹ̀wò nípasẹ̀ `make go-test`, `make go-coverage`, àti `golangci-lint run`. Wo [README èdè Gẹ̀ẹ́sì](../../../README.md) ní kíkún àti [àtọ́ka ìwé-àkọsílẹ̀](../../../docs/README.md). A fún un ní ìwé-àṣẹ lábẹ́ [Apache-2.0](../../../LICENSE).
