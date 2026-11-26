<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licence : Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Configuration répétable de plateforme Kubernetes avec preuves de conformité traçables.

## Aperçu

Mantl est un compilateur de plateforme piloté par des spécifications et un moteur d'exécution de conformité. Une spécification YAML `MantlCluster` est compilée en variables d'entrée Terraform, en manifestes ArgoCD Application et en manifestes d'isolation d'espaces de noms pour les locataires. Un opérateur de conformité Go distinct réconcilie les correspondances de contrôles, les constats, les évaluations et les preuves d'audit S3 versionnées.

- `plan` s'exécute hors ligne pour compiler, enregistrer ou comparer des artefacts de plateforme sans modifier l'infrastructure.
- `apply` exécute des vérifications préalables et amorce les artefacts de plateforme générés ou examinés.
- `status` observe les applications générées et les décomptes de rapports de politiques Kyverno.
- `compliance` inspecte la couverture de conformité, les constats, la remédiation, l'historique et les archives de preuves.

## Démarrage rapide

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Exemple de `platform.yaml` :

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

La CLI `mantl` intègre des traductions pour 45 paramètres régionaux (20 de niveau 1, 15 africains et 10 asiatiques) dans `pkg/i18n`. La sélection du paramètre régional respecte `LC_ALL`, `LC_MESSAGES`, `LANG` ainsi que l'option persistante `--lang` :

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Développement et licence

Exécutez les tests unitaires et les vérifications Go via `make go-test`, `make go-coverage` et `golangci-lint run`. Consultez le [README complet en anglais](../../../README.md) et l'[index de la documentation](../../../docs/README.md). Sous licence [Apache-2.0](../../../LICENSE).
