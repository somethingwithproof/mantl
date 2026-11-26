<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Лиценца: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Поновљива подешавања Kubernetes платформе са следљивим доказима усклађености.

## Преглед

Mantl је компајлер платформе заснован на спецификацији и извршно окружење за усклађеност. YAML спецификација `MantlCluster` компајлира се у улазне променљиве за Terraform, манифесте ArgoCD Application и манифесте изолације простора имена за кориснике. Одвојени Go оператор усклађености усклађује мапирања контрола, налазе, процене и верзионисане ревизијске доказе у S3 складишту.

- `plan` ради ван мреже за компајлирање, чување или поређење артефаката платформе без измена инфраструктуре.
- `apply` извршава предпроверу и покреће генерисане или прегледане артефакте платформе.
- `status` прати генерисане апликације и број извештаја о Kyverno политикама.
- `compliance` прегледа покривеност усклађености, налазе, санацију, историју и архиве доказа.

## Брзи почетак

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Пример датотеке `platform.yaml`:

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
  networking:
    domain: platform.example.com
```

## Локализација

CLI алат `mantl` садржи уграђене преводе за 48 језичких ознака (20 Tier 1, 15 афричких, 10 азијских и 3 јужнословенске) у пакету `pkg/i18n`. Избор језика поштује променљиве `LC_ALL`, `LC_MESSAGES` и `LANG` као и трајну заставицу `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Развој и лиценца

Покрените Go јединичне тестове и провере помоћу `make go-test`, `make go-coverage` и `golangci-lint run`. Погледајте потпуну [енглеску верзију README](../../../README.md) и [индекс документације](../../../docs/README.md). Објављено под лиценцом [Apache-2.0](../../../LICENSE).
