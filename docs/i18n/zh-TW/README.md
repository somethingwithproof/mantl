<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

具備可追蹤合規性證據的可重複 Kubernetes 平台設定工具。

## 概觀

Mantl 是一個規格驅動的平台編譯器與合規性執行階段。`MantlCluster` YAML 規格會編譯為 Terraform 變數輸入、ArgoCD Application 資訊清單以及租戶命名空間隔離資訊清單。獨立的 Go 合規性 Operator 會協調控制項對應、稽核發現、評估結果以及具備版本控制的 S3 稽核證據。

- `plan` 在離線狀態下執行，用於編譯、儲存或比較平台產出物，不會變更基礎架構。
- `apply` 執行事前檢查並啟動已產生或經過審查的平台產出物。
- `status` 觀察已產生的應用程式與 Kyverno 原則報告計數。
- `compliance` 檢視合規性涵蓋範圍、稽核發現、補救措施、歷史紀錄與證據封存檔。

## 快速開始

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

`platform.yaml` 範例：

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

## 在地化

`mantl` CLI 在 `pkg/i18n` 中內嵌了 45 種語系的翻譯（包含 20 種 Tier 1 語系、15 種非洲語系與 10 種亞洲語系）。語系選取會依循 `LC_ALL`、`LC_MESSAGES`、`LANG` 與持續性 `--lang` 旗標：

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## 開發與授權條款

透過 `make go-test`、`make go-coverage` 與 `golangci-lint run` 執行 Go 單元測試與檢查。請參閱完整的[英文 README](../../../README.md)與[文件索引](../../../docs/README.md)。採用 [Apache-2.0](../../../LICENSE) 授權條款。
