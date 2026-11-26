<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

قابلِ تکرار Kubernetes پلیٹ فارم کنفیگریشن، ٹریس ایبل تعمیلی شواہد کے ساتھ۔

## جائزہ

Mantl ایک اسپیک پر مبنی پلیٹ فارم کمپائلر اور کمپلائنس رن ٹائم ہے۔ ایک `MantlCluster` YAML اسپیسفیکیشن Terraform ویری ایبل ان پٹس، ArgoCD Application مینی فیسٹس، اور ٹیننٹ نیم اسپیس آئسولیشن مینی فیسٹس میں کمپائل ہوتی ہے۔ ایک الگ Go کمپلائنس آپریٹر کنٹرول میپنگز، نتائج، جائزوں، اور ورژن شدہ S3 آڈٹ شواہد کو ہم آہنگ کرتا ہے۔

- `plan` بنیادی ڈھانچے میں تبدیلی کیے بغیر پلیٹ فارم آرٹیفیکٹس کو کمپائل کرنے، محفوظ کرنے یا موازنہ کرنے کے لیے آف لائن چلتا ہے۔
- `apply` پری فلائٹ چیکس چلاتا ہے اور تیار کردہ یا نظرثانی شدہ پلیٹ فارم آرٹیفیکٹس کو بوٹ اسٹریپ کرتا ہے۔
- `status` تیار کردہ ایپلیکیشنز اور Kyverno پالیسی رپورٹ کی تعداد کا مشاہدہ کرتا ہے۔
- `compliance` کمپلائنس کوریج، نتائج، تدارک، تاریخچہ اور شواہد کے آرکائیوز کا معائنہ کرتا ہے۔

## فوری آغاز

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

مثال `platform.yaml`:

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

## لوکلائزیشن

`mantl` CLI `pkg/i18n` میں 45 لوکیلز (20 ٹیئر 1، 15 افریقی، اور 10 ایشیائی لوکیلز) کے لیے تراجم شامل کرتا ہے۔ لوکیل کا انتخاب `LC_ALL`، `LC_MESSAGES`، `LANG` اور مستقل `--lang` فلیگ کا احترام کرتا ہے:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## ڈیولپمنٹ اور لائسنس

`make go-test`، `make go-coverage` اور `golangci-lint run` کے ذریعے Go یونٹ ٹیسٹس اور چیکس چلائیں۔ مکمل [انگریزی README](../../../README.md) اور [دستاویزات کی فہرست](../../../docs/README.md) دیکھیں۔ [Apache-2.0](../../../LICENSE) کے تحت لائسنس یافتہ۔
