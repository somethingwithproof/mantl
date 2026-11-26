<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![مجوز: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

پیکربندی تکرارپذیر پلتفرم Kubernetes با شواهد انطباق قابل ردیابی.

## نمای کلی

Mantl یک کامپایلر پلتفرم مبتنی بر مشخصات و زمان اجرای انطباق است. یک مشخصه YAML از نوع `MantlCluster` به ورودی‌های متغیر Terraform، مانیفست‌های Application در ArgoCD و مانیفست‌های ایزوله‌سازی فضای نام مستاجرها کامپایل می‌شود. یک اپراتور انطباق جداگانه به زبان Go، نگاشت‌های کنترلی، یافته‌ها، ارزیابی‌ها و شواهد ممیزی نسخه‌بندی‌شده S3 را همگام‌سازی می‌کند.

- `plan` به‌صورت آفلاین اجرا می‌شود تا آرتیفکت‌های پلتفرم را بدون تغییر در زیرساخت کامپایل، ذخیره یا مقایسه کند.
- `apply` بررسی‌های پیش‌نیاز را انجام می‌دهد و آرتیفکت‌های پلتفرم تولیدشده یا بازبینی‌شده را راه‌اندازی می‌کند.
- `status` برنامه‌های تولیدشده و تعداد گزارش‌های سیاست Kyverno را مشاهده می‌کند.
- `compliance` پوشش انطباق، یافته‌ها، رفع اشکالات، تاریخچه و آرشیوهای شواهد را بررسی می‌کند.

## شروع سریع

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

نمونه `platform.yaml`:

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

## بومی‌سازی

ابزار خط فرمان `mantl` ترجمه‌های ۴۵ زبان محلی (۲۰ زبان رده اول، ۱۵ زبان آفریقایی و ۱۰ زبان آسیایی) را در `pkg/i18n` در خود جای داده است. انتخاب زبان از `LC_ALL`، `LC_MESSAGES`، `LANG` و پرچم ماندگار `--lang` پیروی می‌کند:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## توسعه و مجوز

آزمون‌های واحد و بررسی‌های Go را از طریق `make go-test`، `make go-coverage` و `golangci-lint run` اجرا کنید. برای جزئیات کامل به [نسخه انگلیسی README](../../../README.md) و [فهرست مستندات](../../../docs/README.md) مراجعه کنید. منتشرشده تحت مجوز [Apache-2.0](../../../LICENSE).
