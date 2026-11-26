<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Повторяемая конфигурация платформы Kubernetes с отслеживаемыми доказательствами соответствия требованиям.

## Обзор

Mantl представляет собой декларативный компилятор платформы на основе спецификаций и среду исполнения для контроля соответствия требованиям. YAML-спецификация `MantlCluster` компилируется во входные переменные Terraform, манифесты ArgoCD Application и манифесты изоляции пространств имён арендаторов. Отдельный Go-оператор соответствия согласует сопоставления контролей, выявленные несоответствия, оценки и версионированные аудиторские свидетельства в S3.

- `plan` выполняется офлайн для компиляции, сохранения или сравнения артефактов платформы без изменения инфраструктуры.
- `apply` выполняет предварительные проверки и инициализирует сгенерированные или проверенные артефакты платформы.
- `status` отслеживает сгенерированные приложения и счётчики отчётов политик Kyverno.
- `compliance` проверяет охват требований соответствия, выявленные несоответствия, меры по устранению, историю и архивы доказательств.

## Быстрый старт

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Пример `platform.yaml`:

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

## Локализация

CLI `mantl` включает встроенные переводы для 45 локалей (20 локалей Tier 1, 15 африканских и 10 азиатских) в `pkg/i18n`. Выбор локали учитывает `LC_ALL`, `LC_MESSAGES`, `LANG` и постоянный флаг `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Разработка и лицензия

Запускайте модульные тесты и проверки Go с помощью `make go-test`, `make go-coverage` и `golangci-lint run`. Подробнее см. в полной [английской версии README](../../../README.md) и [указателе документации](../../../docs/README.md). Распространяется под лицензией [Apache-2.0](../../../LICENSE).
