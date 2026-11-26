<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Usanidi unaorudiwa wa jukwaa la Kubernetes wenye ushahidi wa uzingatiaji unaoweza kufuatiliwa.

## Muhtasari

Mantl ni kikusanyaji cha jukwaa kinachoongozwa na vipimo na mazingira ya utekelezaji wa uzingatiaji. Vipimo vya YAML vya `MantlCluster` hukusanywa kuwa ingizo za vigezo vya Terraform, faili za maelezo za ArgoCD Application, na faili za kutenganisha namespace za wapangaji. Opereta tofauti wa uzingatiaji wa Go hulinganisha mipangilio ya udhibiti, matokeo ya ukaguzi, tathmini, na ushahidi wa ukaguzi wa S3 wenye matoleo.

- `plan` huendeshwa nje ya mtandao ili kukusanya, kuhifadhi, au kulinganisha vijenzi vya jukwaa bila kubadilisha miundombinu.
- `apply` huendesha ukaguzi wa awali na kuanzisha vijenzi vya jukwaa vilivyoundwa au kukaguliwa.
- `status` hufuatilia programu zilizoundwa na idadi ya ripoti za sera za Kyverno.
- `compliance` hukagua kiwango cha uzingatiaji, matokeo, urekebishaji, historia, na hifadhi za ushahidi.

## Mwanzo wa haraka

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Mfano wa `platform.yaml`:

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

## Ujanibishaji

CLI ya `mantl` hujumuisha tafsiri za lugha na maeneo 45 (maeneo 20 ya Daraja la 1, 15 ya Afrika, na 10 ya Asia) katika `pkg/i18n`. Uteuzi wa eneo huheshimu `LC_ALL`, `LC_MESSAGES`, `LANG`, na bendera ya kudumu ya `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Uendelezaji na Leseni

Endesha majaribio ya kitengo ya Go na ukaguzi kupitia `make go-test`, `make go-coverage`, na `golangci-lint run`. Tazama [README kamili ya Kiingereza](../../../README.md) na [faharasa ya nyaraka](../../../docs/README.md). Imepewa leseni chini ya [Apache-2.0](../../../LICENSE).
