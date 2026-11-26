<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licencia: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Configuración repetible de plataformas Kubernetes con evidencia de cumplimiento trazable.

## Descripción general

Mantl es un compilador de plataformas basado en especificaciones y un entorno de ejecución de cumplimiento. Una especificación YAML `MantlCluster` se compila en entradas de variables de Terraform, manifiestos Application de ArgoCD y manifiestos de aislamiento de espacios de nombres de inquilinos. Un operador de cumplimiento en Go independiente reconcilia asignaciones de controles, hallazgos, evaluaciones y evidencia de auditoría versionada en S3.

- `plan` se ejecuta sin conexión para compilar, guardar o comparar artefactos de la plataforma sin modificar la infraestructura.
- `apply` ejecuta comprobaciones previas e inicializa los artefactos de plataforma generados o revisados.
- `status` observa las aplicaciones generadas y los recuentos de informes de políticas de Kyverno.
- `compliance` inspecciona la cobertura de cumplimiento, los hallazgos, la remediación, el historial y los archivos de evidencia.

## Inicio rápido

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Ejemplo de `platform.yaml`:

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

## Localización

La CLI `mantl` integra traducciones para 45 configuraciones regionales (20 de nivel 1, 15 africanas y 10 asiáticas) en `pkg/i18n`. La selección de configuración regional respeta `LC_ALL`, `LC_MESSAGES`, `LANG` y el indicador persistente `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Desarrollo y licencia

Ejecuta las pruebas unitarias y comprobaciones de Go mediante `make go-test`, `make go-coverage` y `golangci-lint run`. Consulta el [README completo en inglés](../../../README.md) y el [índice de documentación](../../../docs/README.md). Publicado bajo la licencia [Apache-2.0](../../../LICENSE).
