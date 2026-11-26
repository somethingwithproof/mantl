<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

إعدادات منصة Kubernetes قابلة للتكرار مع أدلة امتثال (compliance evidence) قابلة للتتبع.

## نظرة عامة

Mantl عبارة عن platform compiler وبيئة تشغيل للامتثال بتعتمد على المواصفات (spec-driven). ملف مواصفات `MantlCluster` بصيغة YAML بيتحوّل لمدخلات متغيرات Terraform، وملفات ArgoCD Application manifests، وملفات عزل الـ namespaces للمستأجرين (tenants). وفيه compliance operator منفصل مكتوب بلغة Go بيعمل مزامنة لربط الضوابط (control mappings)، والنتائج (findings)، والتقييمات، وأدلة التدقيق المؤرشفة بإصدارات على S3.

- `plan` بيشتغل أوفلاين عشان يجمع أو يحفظ أو يقارن مخرجات المنصة (artifacts) من غير ما يغيّر في البنية التحتية.
- `apply` بينفّذ فحوصات ما قبل التشغيل (preflight checks) وبيجهّز مخرجات المنصة المولّدة أو اللي اتراجعت.
- `status` بيتابع التطبيقات المولّدة وإحصائيات تقارير سياسات Kyverno.
- `compliance` بيفحص تغطية الامتثال، والنتائج، والمعالجة (remediation)، والسجل التاريخي، وأرشيف الأدلة.

## البداية السريعة

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

## اللغات والترجمة

أداة `mantl` مدمج جواها ترجمات لـ 45 لغة وإعداد محلي (20 من الفئة الأولى Tier 1، و15 لغة أفريقية، و10 لغات آسيوية) في `pkg/i18n`. اختيار اللغة بيتبع `LC_ALL` و`LC_MESSAGES` و`LANG`، وكمان خيار `--lang` الدائم:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## التطوير والترخيص

شغّل اختبارات الوحدة وفحوصات Go عن طريق `make go-test` و`make go-coverage` و`golangci-lint run`. شوف [ملف README الإنجليزي](../../../README.md) الكامل و[فهرس التوثيق](../../../docs/README.md). مرخّص بموجب [Apache-2.0](../../../LICENSE).
