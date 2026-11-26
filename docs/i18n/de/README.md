<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lizenz: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Reproduzierbare Kubernetes-Plattformkonfiguration mit nachvollziehbaren Compliance-Nachweisen.

## Überblick

Mantl ist ein spezifikationsgesteuerter Plattform-Compiler und eine Compliance-Laufzeitumgebung. Eine `MantlCluster`-YAML-Spezifikation wird in Terraform-Variableneingaben, ArgoCD-Application-Manifeste und Manifeste zur Mandanten-Namespace-Isolierung kompiliert. Ein separater Go-Compliance-Operator gleicht Kontrollzuordnungen, Befunde, Auswertungen und versionierte S3-Audit-Nachweise ab.

- `plan` wird offline ausgeführt, um Plattformartefakte zu kompilieren, zu speichern oder zu vergleichen, ohne die Infrastruktur zu verändern.
- `apply` führt Vorabprüfungen durch und initialisiert generierte oder geprüfte Plattformartefakte.
- `status` überwacht generierte Anwendungen und die Anzahl der Kyverno-Richtlinienberichte.
- `compliance` prüft Compliance-Abdeckung, Befunde, Behebungsmaßnahmen, Verlauf und Nachweisarchive.

## Schnellstart

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Beispiel für `platform.yaml`:

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

## Lokalisierung

Die `mantl`-CLI bettet Übersetzungen für 45 Locales (20 Tier-1-, 15 afrikanische und 10 asiatische Locales) in `pkg/i18n` ein. Die Locale-Auswahl berücksichtigt `LC_ALL`, `LC_MESSAGES`, `LANG` sowie das persistente Flag `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Entwicklung und Lizenz

Go-Unit-Tests und Prüfungen mit `make go-test`, `make go-coverage` und `golangci-lint run` ausführen. Weitere Details finden sich in der vollständigen [englischen README](../../../README.md) und im [Dokumentationsindex](../../../docs/README.md). Lizenziert unter [Apache-2.0](../../../LICENSE).
