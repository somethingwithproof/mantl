<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licence : Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Configuration reproductible de plateforme Kubernetes avec preuves de conformité traçables.

## Aperçu

Mantl est un compilateur de plateforme piloté par spécification et un environnement d'exécution de conformité. Une spécification YAML `MantlCluster` est compilée en variables d'entrée Terraform, en manifestes ArgoCD Application et en manifestes d'isolation d'espaces de noms pour les locataires. Un opérateur Go distinct dédié à la conformité réconcilie les correspondances de contrôles, les constats, les évaluations et les preuves d'audit versionnées sur S3.

- `plan` fonctionne hors ligne pour compiler, enregistrer ou comparer les artefacts de plateforme sans modifier l'infrastructure.
- `apply` exécute les vérifications préalables et amorce les artefacts de plateforme générés ou relus.
- `status` observe les applications générées et les compteurs de rapports de politiques Kyverno.
- `compliance` inspecte la couverture de conformité, les constats, les remédiations, l'historique et les archives de preuves.

## Démarrage rapide

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Exemple de fichier `platform.yaml` :

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

## Localisation

La CLI `mantl` embarque des traductions pour 45 locales (20 de niveau 1, 15 africaines et 10 asiatiques) dans `pkg/i18n`. La sélection de la locale respecte `LC_ALL`, `LC_MESSAGES` et `LANG`, ainsi que l'option persistante `--lang` :

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Développement et licence

Exécutez les tests unitaires et vérifications Go via `make go-test`, `make go-coverage` et `golangci-lint run`. Consultez le [README anglais](../../../README.md) complet et l'[index de la documentation](../../../docs/README.md). Distribué sous licence [Apache-2.0](../../../LICENSE).
