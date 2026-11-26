<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![الترخيص: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

إعدادات قابلة للتكرار لمنصة Kubernetes مع أدلة امتثال قابلة للتتبع.

## نظرة عامة

Mantl هو مترجم منصات قائم على المواصفات وبيئة تشغيل للامتثال. تترجم مواصفة YAML من نوع `MantlCluster` إلى مدخلات متغيرات Terraform، وملفات ArgoCD Application، وملفات عزل مساحات الأسماء للمستأجرين. ويتولى مشغل امتثال منفصل بلغة Go مطابقة الضوابط والنتائج والتقييمات وأدلة التدقيق ذات الإصدارات على S3.

- يعمل `plan` دون اتصال بالشبكة لترجمة مخرجات المنصة أو حفظها أو مقارنتها دون تعديل البنية التحتية.
- يجري `apply` فحوصات ما قبل التشغيل ويهيئ مخرجات المنصة المولدة أو المراجعة.
- يراقب `status` التطبيقات المولدة وعدد تقارير سياسات Kyverno.
- يفحص `compliance` تغطية الامتثال والنتائج والمعالجة والسجل وأرشيفات الأدلة.

## البدء السريع

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

مثال لملف `platform.yaml`:

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

## التوطين

تدمج أداة سطر الأوامر `mantl` ترجمات لـ 45 لغة محلية (20 من المستوى 1، و15 لغة إفريقية، و10 لغات آسيوية) داخل `pkg/i18n`. يحترم اختيار اللغة متغيرات `LC_ALL` و`LC_MESSAGES` و`LANG`، بالإضافة إلى الخيار الدائم `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## التطوير والترخيص

شغل اختبارات الوحدة وفحوصات Go عبر `make go-test` و`make go-coverage` و`golangci-lint run`. راجع [ملف README الإنجليزي الكامل](../../../README.md) و[فهرس التوثيق](../../../docs/README.md). مرخص بموجب [Apache-2.0](../../../LICENSE).
