<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![许可证：Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

可重复的 Kubernetes 平台配置，提供可追溯的合规性证据。

## 概述

Mantl 是一个规格驱动的平台编译器与合规运行时。`MantlCluster` YAML 规格文件可编译为 Terraform 变量输入、ArgoCD Application 清单以及租户命名空间隔离清单。独立的 Go 合规性 Operator 负责协调控制映射、发现项、评估结果以及带版本控制的 S3 审计证据。

- `plan` 离线运行，用于编译、保存或比较平台产物，而不会更改基础设施。
- `apply` 运行预检检查，并引导生成或已审阅的平台产物。
- `status` 观测已生成的应用程序及 Kyverno 策略报告计数。
- `compliance` 检查合规覆盖率、发现项、修复措施、历史记录和证据归档。

## 快速开始

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

`platform.yaml` 示例：

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

## 本地化

`mantl` CLI 在 `pkg/i18n` 中内置了 45 个语言环境的翻译（20 个 Tier 1、15 个非洲和 10 个亚洲语言环境）。语言环境选择遵循 `LC_ALL`、`LC_MESSAGES`、`LANG` 以及持久标志 `--lang`：

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## 开发与许可证

通过 `make go-test`、`make go-coverage` 和 `golangci-lint run` 运行 Go 单元测试与代码检查。请参阅完整的[英文 README](../../../README.md) 和[文档索引](../../../docs/README.md)。基于 [Apache-2.0](../../../LICENSE) 协议开源。
