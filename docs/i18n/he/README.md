<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![רישיון: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

תצורת פלטפורמת Kubernetes הדירה עם ראיות תאימות הניתנות למעקב.

## סקירה כללית

Mantl הוא מהדר פלטפורמה מונחה-מפרט וסביבת ריצה לתאימות. מפרט YAML מסוג `MantlCluster` עובר הידור לקלטי משתנים של Terraform, למניפסטים של ArgoCD Application ולמניפסטים לבידוד מרחבי שמות (namespaces) של דיירים. אופרטור תאימות נפרד ב-Go מסנכרן מיפויי בקרה, ממצאים, הערכות וראיות ביקורת בעלות גרסאות ב-S3.

- `plan` פועל במצב לא מקוון כדי להדר, לשמור או להשוות תוצרי פלטפורמה מבלי לשנות תשתית.
- `apply` מריץ בדיקות מקדימות ומאתחל תוצרי פלטפורמה שנוצרו או נסקרו.
- `status` עוקב אחר יישומים שנוצרו וספירות של דוחות מדיניות Kyverno.
- `compliance` בוחן כיסוי תאימות, ממצאים, תיקון, היסטוריה וארכיוני ראיות.

## התחלה מהירה

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

דוגמה לקובץ `platform.yaml`:

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

## לוקליזציה

כלי ה-CLI של `mantl` מטמיע תרגומים עבור 45 אזורי שפה (20 אזורי Tier 1, 15 אזורים אפריקאיים ו-10 אזורים אסייתיים) תחת `pkg/i18n`. בחירת אזור השפה מתחשבת במשתנים `LC_ALL`, `LC_MESSAGES`, `LANG` ובדגל המתמיד `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## פיתוח ורישיון

הריצו בדיקות יחידה ובדיקות Go באמצעות `make go-test`, `make go-coverage` ו-`golangci-lint run`. ראו את קובץ ה-[README המלא באנגלית](../../../README.md) ואת [אינדקס התיעוד](../../../docs/README.md). מורשה תחת [Apache-2.0](../../../LICENSE).
