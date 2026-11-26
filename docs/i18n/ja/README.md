<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

追跡可能なコンプライアンス証跡を備えた、再現性の高い Kubernetes プラットフォーム設定ツールです。

## 概要

Mantl は仕様駆動のプラットフォームコンパイラ兼コンプライアンスランタイムです。`MantlCluster` YAML 仕様を Terraform 変数入力、ArgoCD Application マニフェスト、テナント名前空間分離マニフェストにコンパイルします。また、独立した Go 製コンプライアンスオペレーターが、コントロールマッピング、検出結果、評価結果、バージョン管理された S3 監査証跡を調整（reconcile）します。

- `plan`: インフラストラクチャを変更せずにオフラインで実行し、プラットフォーム成果物のコンパイル、保存、比較を行います。
- `apply`: 事前チェックを実行し、生成またはレビュー済みのプラットフォーム成果物をブートストラップします。
- `status`: 生成されたアプリケーションおよび Kyverno ポリシーレポートの件数を監視します。
- `compliance`: コンプライアンスのカバー状況、検出結果、修復状況、履歴、および証跡アーカイブを確認します。

## クイックスタート

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

`platform.yaml` の例:

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

## ローカライズ

`mantl` CLI には、45 ロケール（Tier 1 が 20、アフリカ地域が 15、アジア地域が 10）の翻訳が `pkg/i18n` に埋め込まれています。ロケール選択は `LC_ALL`、`LC_MESSAGES`、`LANG`、および永続フラグ `--lang` を参照します。

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## 開発とライセンス

Go の単体テストおよび各種チェックは `make go-test`、`make go-coverage`、`golangci-lint run` で実行できます。詳細は[英語版 README](../../../README.md)および[ドキュメント索引](../../../docs/README.md)を参照してください。ライセンスは [Apache-2.0](../../../LICENSE) です。
