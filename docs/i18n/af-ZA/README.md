<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lisensie: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Herhaalbare Kubernetes-platformkonfigurasie met naspeurbare nakomingsbewyse.

## Oorsig

Mantl is 'n spesifikasiegedrewe platformsamesteller en nakomingslooptydomgewing. 'n `MantlCluster` YAML-spesifikasie word saamgestel in Terraform-veranderlike-insette, ArgoCD Application-manifeste en huurder-naamruimte-isolasiemanifeste. 'n Afsonderlike Go-nakomingsoperator versoen kontroletoewysings, bevindings, evaluasies en weergawebeheerde S3-ouditbewyse.

- `plan` loop vanlyn om platformartefakte saam te stel, te stoor of te vergelyk sonder om infrastruktuur te verander.
- `apply` voer voorafgaande kontroles uit en inisialiseer gegenereerde of hersiende platformartefakte.
- `status` neem gegenereerde toepassings en Kyverno-beleidsverslagtellings waar.
- `compliance` inspekteer nakomingsdekking, bevindings, remediëring, geskiedenis en bewysargiewe.

## Vinnige begin

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Voorbeeld van `platform.yaml`:

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

## Lokalisering

Die `mantl`-CLI sluit vertalings vir 45 locales (20 Vlak 1-, 15 Afrika- en 10 Asiatiese locales) in `pkg/i18n` in. Locale-keuse respekteer `LC_ALL`, `LC_MESSAGES` en `LANG`, sowel as die blywende `--lang`-vlag:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Ontwikkeling en lisensie

Voer Go-eenheidstoetse en -kontroles uit via `make go-test`, `make go-coverage` en `golangci-lint run`. Sien die volledige [Engelse README](../../../README.md) en [dokumentasie-indeks](../../../docs/README.md). Gelisensieer onder [Apache-2.0](../../../LICENSE).
