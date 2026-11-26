<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Fikirakirana sehatra Kubernetes azo averina miaraka amin'ny porofo fanarahan-dalàna azo arahina.

## Fijerena ankapobeny

Ny Mantl dia mpanangona sehatra miorina amin'ny famaritana (spec) ary mpandrindra fanarahan-dalàna. Ny famaritana YAML `MantlCluster` dia atambatra ho fampidirana mari-ova Terraform, manifest ArgoCD Application, ary manifest fanavahana ny namespace an'ny mpampiasa. Ny mpandraharaha fanarahan-dalàna Go misaraka dia mampifanaraka ny fampifandraisana fanaraha-maso, ny zavatra hita, ny tombana, ary ny porofo fanamarinana S3 misy dikan-teny.

- `plan` dia mandeha tsy misy tambajotra mba hanangona, hitahiry, na hampitaha ireo vokatra ara-tsehatra nefa tsy manova ny fotodrafitrasa.
- `apply` dia manao fanamarinana mialoha ary manangana ireo vokatra ara-tsehatra namboarina na nojerena.
- `status` dia manara-maso ireo rindranasa namboarina sy ny isan'ny tatitra momba ny fitsipika Kyverno.
- `compliance` dia mandinika ny fandrakofana ny fanarahan-dalàna, ny zavatra hita, ny fanarenana, ny tantara, ary ny tahiry porofo.

## Fanombohana haingana

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Ohatra `platform.yaml`:

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

## Fampifanarahana amin'ny fiteny

Ny CLI `mantl` dia ahitana fandikan-teny ho an'ny fiteny eo an-toerana miisa 45 (fiteny sokajy 1 miisa 20, fiteny afrikanina miisa 15, ary fiteny aziatika miisa 10) ao amin'ny `pkg/i18n`. Ny fisafidianana ny fiteny dia manaja ny `LC_ALL`, `LC_MESSAGES`, `LANG`, ary ny saina maharitra `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Fampandrosoana sy Lisansa

Alefaso ny fitsapana tarika Go sy ny fanamarinana amin'ny alalan'ny `make go-test`, `make go-coverage`, ary `golangci-lint run`. Jereo ny [README amin'ny teny anglisy](../../../README.md) feno sy ny [lisitry ny tahirin-kevitra](../../../docs/README.md). Navoaka teo ambany lisansa [Apache-2.0](../../../LICENSE).
