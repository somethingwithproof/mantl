<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licentie: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Herhaalbare Kubernetes-platformconfiguratie met traceerbaar compliancebewijs.

## Overzicht

Mantl is een op specificaties gebaseerde platformcompiler en compliance-runtime. Een `MantlCluster` YAML-specificatie wordt gecompileerd naar Terraform-variabele-invoer, ArgoCD Application-manifesten en isolatiemanifesten voor tenant-namespaces. Een afzonderlijke Go compliance-operator reconcilieert koppelingen van beheersmaatregelen, bevindingen, evaluaties en geversioneerd S3-auditbewijs.

- `plan` draait offline om platformartefacten te compileren, op te slaan of te vergelijken zonder de infrastructuur te wijzigen.
- `apply` voert preflight-controles uit en bootstrapt gegenereerde of gecontroleerde platformartefacten.
- `status` observeert gegenereerde applicaties en aantallen Kyverno-beleidsrapporten.
- `compliance` inspecteert compliance-dekking, bevindingen, herstelmaatregelen, geschiedenis en bewijsarchieven.

## Snelstart

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

## Lokalisatie

De `mantl`-CLI bevat ingebouwde vertalingen voor 45 locales (20 Tier 1-, 15 Afrikaanse en 10 Aziatische locales) in `pkg/i18n`. De selectie van de locale respecteert `LC_ALL`, `LC_MESSAGES`, `LANG` en de persistente vlag `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Ontwikkeling en licentie

Voer Go-unittests en controles uit via `make go-test`, `make go-coverage` en `golangci-lint run`. Zie de volledige [Engelse README](../../../README.md) en [documentatie-index](../../../docs/README.md). Gelicentieerd onder [Apache-2.0](../../../LICENSE).
