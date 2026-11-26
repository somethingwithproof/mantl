<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Ilayisensi: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Ubumbeko lweqonga le-Kubernetes oluphindaphindekayo olunobungqina bokuthobela imigaqo obulandelelekayo.

## Ushwankathelo

I-Mantl ngumqulunqi weqonga oqhutywa yinkcazo kunye nexesha lokuqhuba lokuthobela imigaqo. Inkcazo ye-YAML ye-`MantlCluster` iqulunqwa ibe zizinto zokufaka eziguquguqukayo ze-Terraform, iimanifest ze-ArgoCD Application, kunye neemanifest zokwahlula ii-namespace zabaqeshi. Umsebenzi we-Go wokuthobela imigaqo owahlukileyo ungqamanisa iimephu zolawulo, iziphumo, uvavanyo, kunye nobungqina bophicotho-zincwadi be-S3 obuneenguqulelo.

- `plan` isebenza ngaphandle kwe-intanethi ukuqulunqa, ukugcina, okanye ukuthelekisa izinto zeqonga ngaphandle kokutshintsha iziseko zophuhliso.
- `apply` iqhuba iitshekhi zangaphambili kwaye iqalise izinto zeqonga ezivelisiweyo okanye eziphononongiweyo.
- `status` ijonga usetyenziso oluvelisiweyo kunye namanani eengxelo zomgaqo-nkqubo we-Kyverno.
- `compliance` ihlola ukugubungela kokuthobela imigaqo, iziphumo, ukulungiswa, imbali, kunye noovimba bobungqina.

## Ukuqala ngokukhawuleza

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Umzekelo we-`platform.yaml`:

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

## Ulungiselelo lweelwimi

I-`mantl` CLI ifakela iinguqulelo zeelwimi zasekuhlaleni ezingama-45 (ezingama-20 ze-Tier 1, ezili-15 zase-Afrika, nezili-10 zase-Asia) kwi-`pkg/i18n`. Ukukhethwa kolwimi lwasekuhlaleni kuhlonipha i-`LC_ALL`, `LC_MESSAGES`, `LANG`, kunye nefleki eqhubekayo ye-`--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Uphuhliso neLayisensi

Qhuba iimvavanyo zeyunithi ze-Go kunye neetshekhi usebenzisa `make go-test`, `make go-coverage`, kunye `golangci-lint run`. Jonga i-[README yesiNgesi](../../../README.md) epheleleyo kunye [nesalathiso samaxwebhu](../../../docs/README.md). Inikwe ilayisensi phantsi kwe-[Apache-2.0](../../../LICENSE).
