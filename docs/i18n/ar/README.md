<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

إعدادات منصّة Kubernetes قابلة للتكرار مع أدلّة امتثال قابلة للتتبّع.

## نظرة عامة

`Mantl` هو مُصرِّف منصّة قائم على المواصفات وبيئة تشغيل للامتثال. تُصرَّف مواصفة `MantlCluster` بصيغة YAML إلى مُدخلات متغيّرات Terraform، وبيانات تطبيقات ArgoCD، وبيانات عزل مساحات أسماء المستأجرين. ويتولّى مُشغِّل امتثال منفصل بلغة Go مطابقة تعيينات الضوابط، والنتائج، والتقييمات، وأدلّة التدقيق ذات الإصدارات على S3.

- يعمل `plan` دون اتصال بالإنترنت لتصريف مخرجات المنصّة أو حفظها أو مقارنتها دون تعديل البنية التحتية.
- يُجري `apply` فحوصات ما قبل التشغيل ويهيّئ مخرجات المنصّة المُولَّدة أو المُراجَعة.
- يراقب `status` التطبيقات المُولَّدة وأعداد تقارير سياسات Kyverno.
- يفحص `compliance` تغطية الامتثال، والنتائج، والمعالجة، والسجل التاريخي، وأرشيفات الأدلّة.

## البدء السريع

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

مثال على `platform.yaml`:

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

تُضمِّن أداة سطر الأوامر `mantl` ترجمات لـ 45 لغةً وإعدادًا محليًا (20 لغةً من المستوى الأول، و15 لغةً أفريقيةً، و10 لغاتٍ آسيويةٍ) في `pkg/i18n`. يحترم اختيار اللغة متغيّرات `LC_ALL` و`LC_MESSAGES` و`LANG`، بالإضافة إلى الراية الدائمة `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## التطوير والترخيص

شغّل اختبارات الوحدة وفحوصات Go عبر `make go-test` و`make go-coverage` و`golangci-lint run`. راجع [ملف README الإنجليزي الكامل](../../../README.md) و[فهرس التوثيق](../../../docs/README.md). مرخّص بموجب [Apache-2.0](../../../LICENSE).
