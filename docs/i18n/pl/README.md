<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licencja: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Powtarzalna konfiguracja platformy Kubernetes ze śledzonymi dowodami zgodności.

## Przegląd

Mantl to sterowany specyfikacją kompilator platformy i środowisko uruchomieniowe zgodności. Specyfikacja YAML `MantlCluster` kompiluje się do zmiennych wejściowych Terraform, manifestów ArgoCD Application oraz manifestów izolacji przestrzeni nazw najemców. Osobny operator zgodności w języku Go uzgadnia mapowania mechanizmów kontrolnych, wykryte niezgodności, oceny oraz wersjonowane dowody audytowe w S3.

- `plan` działa w trybie offline, aby kompilować, zapisywać lub porównywać artefakty platformy bez wprowadzania zmian w infrastrukturze.
- `apply` wykonuje wstępne testy kontrolne i inicjuje wygenerowane lub zweryfikowane artefakty platformy.
- `status` monitoruje wygenerowane aplikacje i liczbę raportów polityk Kyverno.
- `compliance` analizuje pokrycie zgodności, wykryte niezgodności, działania naprawcze, historię oraz archiwa dowodów.

## Szybki start

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Przykładowy plik `platform.yaml`:

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

## Lokalizacja

Narzędzie CLI `mantl` zawiera wbudowane tłumaczenia dla 45 ustawień regionalnych (20 z grupy Tier 1, 15 afrykańskich i 10 azjatyckich) w pakiecie `pkg/i18n`. Wybór języka uwzględnia zmienne `LC_ALL`, `LC_MESSAGES`, `LANG` oraz trwałą flagę `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Rozwój i licencja

Uruchamiaj testy jednostkowe i kontrole w Go za pomocą `make go-test`, `make go-coverage` oraz `golangci-lint run`. Zobacz pełną [angielską wersję README](../../../README.md) oraz [indeks dokumentacji](../../../docs/README.md). Udostępniono na licencji [Apache-2.0](../../../LICENSE).
