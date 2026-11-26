<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Ліцензія: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Відтворювана конфігурація платформи Kubernetes із відстежуваними доказами відповідності вимогам.

## Огляд

Mantl є керованим специфікаціями компілятором платформи та середовищем виконання для контролю відповідності. YAML-специфікація `MantlCluster` компілюється у вхідні змінні Terraform, маніфести ArgoCD Application та маніфести ізоляції просторів імен орендарів. Окремий Go-оператор відповідності узгоджує зіставлення контролів, виявлені невідповідності, оцінювання та версіоновані аудиторські докази в S3.

- `plan` працює в автономному режимі для компіляції, збереження або порівняння артефактів платформи без зміни інфраструктури.
- `apply` виконує попередні перевірки та розгортає згенеровані або перевірені артефакти платформи.
- `status` відстежує згенеровані застосунки та кількість звітів політик Kyverno.
- `compliance` перевіряє покриття відповідності вимогам, виявлені невідповідності, усунення проблем, історію та архіви доказів.

## Швидкий старт

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Приклад `platform.yaml`:

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

## Локалізація

CLI `mantl` містить вбудовані переклади для 45 локалей (20 рівня Tier 1, 15 африканських і 10 азійських локалей) у `pkg/i18n`. Вибір локалі враховує `LC_ALL`, `LC_MESSAGES`, `LANG` та постійний прапорець `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Розробка та ліцензія

Запускайте модульні тести Go та перевірки за допомогою `make go-test`, `make go-coverage` та `golangci-lint run`. Див. повний [README англійською](../../../README.md) та [покажчик документації](../../../docs/README.md). Ліцензовано за [Apache-2.0](../../../LICENSE).
