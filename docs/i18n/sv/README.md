<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licens: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Repeterbar Kubernetes-plattformskonfiguration med spårbara efterlevnadsbevis.

## Översikt

Mantl är en specifikationsdriven plattformskompilator och körtidsmiljö för regelefterlevnad. En `MantlCluster`-specifikation i YAML kompileras till Terraform-variabelindata, ArgoCD Application-manifest och manifest för isolering av tenant-namnrymder. En separat Go-baserad efterlevnadsoperator stämmer av kontrollmappningar, fynd, utvärderingar och versionshanterade S3-revisionsbevis.

- `plan` körs offline för att kompilera, spara eller jämföra plattformsartefakter utan att ändra infrastrukturen.
- `apply` kör förkontroller och initierar genererade eller granskade plattformsartefakter.
- `status` övervakar genererade applikationer och antal Kyverno-policyrapporter.
- `compliance` inspekterar efterlevnadstäckning, fynd, åtgärder, historik och bevisarkiv.

## Snabbstart

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Exempel på `platform.yaml`:

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

CLI-verktyget `mantl` bäddar in översättningar för 45 språkområden (20 Tier 1-, 15 afrikanska och 10 asiatiska språkområden) i `pkg/i18n`. Språkvalet följer `LC_ALL`, `LC_MESSAGES`, `LANG` och den beständiga flaggan `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Utveckling och licens

Kör enhetstester och kontroller i Go med `make go-test`, `make go-coverage` och `golangci-lint run`. Se den fullständiga [engelska README-filen](../../../README.md) och [dokumentationsindexet](../../../docs/README.md). Licensierad under [Apache-2.0](../../../LICENSE).
