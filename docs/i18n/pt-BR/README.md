<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Configuração reproduzível de plataforma Kubernetes com evidências rastreáveis de conformidade.

## Visão geral

Mantl é um compilador de plataforma orientado a especificações e um runtime de conformidade. Uma especificação YAML `MantlCluster` é compilada em entradas de variáveis do Terraform, manifestos Application do ArgoCD e manifestos de isolamento de namespaces de tenants. Um operador de conformidade separado em Go reconcilia mapeamentos de controles, achados, avaliações e evidências de auditoria versionadas no S3.

- `plan` é executado offline para compilar, salvar ou comparar artefatos da plataforma sem alterar a infraestrutura.
- `apply` executa verificações prévias e faz o bootstrap dos artefatos da plataforma gerados ou revisados.
- `status` observa as aplicações geradas e as contagens de relatórios de políticas do Kyverno.
- `compliance` inspeciona a cobertura de conformidade, achados, remediação, histórico e arquivos de evidências.

## Início rápido

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Exemplo de `platform.yaml`:

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

## Localização

A CLI `mantl` incorpora traduções para 45 locales (20 Tier 1, 15 africanos e 10 asiáticos) em `pkg/i18n`. A seleção de locale respeita `LC_ALL`, `LC_MESSAGES`, `LANG` e a flag persistente `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Desenvolvimento e Licença

Execute testes unitários e verificações em Go com `make go-test`, `make go-coverage` e `golangci-lint run`. Consulte o [README em inglês](../../../README.md) completo e o [índice de documentação](../../../docs/README.md). Licenciado sob [Apache-2.0](../../../LICENSE).
