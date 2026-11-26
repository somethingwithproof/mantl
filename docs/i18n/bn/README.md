<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

অনুসরণযোগ্য কমপ্লায়েন্স প্রমাণসহ পুনরাবৃত্তিযোগ্য Kubernetes প্ল্যাটফর্ম কনফিগারেশন।

## সংক্ষিপ্ত বিবরণ

Mantl হলো একটি স্পেক-চালিত প্ল্যাটফর্ম কম্পাইলার এবং কমপ্লায়েন্স রানটাইম। একটি `MantlCluster` YAML স্পেসিফিকেশন Terraform ভেরিয়েবল ইনপুট, ArgoCD Application ম্যানিফেস্ট এবং টেন্যান্ট নেমস্পেস আইসোলেশন ম্যানিফেস্টে কম্পাইল হয়। একটি পৃথক Go কমপ্লায়েন্স অপারেটর কন্ট্রোল ম্যাপিং, ফাইন্ডিংস, মূল্যায়ন এবং ভার্সন করা S3 অডিট প্রমাণ সমন্বয় করে।

- `plan` অবকাঠামো পরিবর্তন না করেই প্ল্যাটফর্ম আর্টিফ্যাক্ট কম্পাইল, সংরক্ষণ বা তুলনা করতে অফলাইনে চলে।
- `apply` প্রিফ্লাইট পরীক্ষা চালায় এবং উৎপন্ন বা পর্যালোচিত প্ল্যাটফর্ম আর্টিফ্যাক্টগুলো বুটস্ট্র্যাপ করে।
- `status` উৎপন্ন অ্যাপ্লিকেশন এবং Kyverno পলিসি রিপোর্টের সংখ্যা পর্যবেক্ষণ করে।
- `compliance` কমপ্লায়েন্স কভারেজ, ফাইন্ডিংস, প্রতিকার, ইতিহাস এবং প্রমাণ আর্কাইভ পরীক্ষা করে।

## দ্রুত শুরু

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

উদাহরণ `platform.yaml`:

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

## স্থানীয়করণ

`mantl` CLI `pkg/i18n`-এ ৪৫টি লোকেলের (২০টি Tier 1, ১৫টি আফ্রিকান এবং ১০টি এশীয় লোকেল) অনুবাদ এম্বেড করে। লোকেল নির্বাচন `LC_ALL`, `LC_MESSAGES`, `LANG` এবং স্থায়ী `--lang` ফ্ল্যাগ অনুসরণ করে:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## উন্নয়ন ও লাইসেন্স

`make go-test`, `make go-coverage` এবং `golangci-lint run`-এর মাধ্যমে Go ইউনিট টেস্ট এবং পরীক্ষাগুলো চালান। সম্পূর্ণ [ইংরেজি README](../../../README.md) এবং [ডকুমেন্টেশন সূচিপত্র](../../../docs/README.md) দেখুন। [Apache-2.0](../../../LICENSE) লাইসেন্সের অধীনে লাইসেন্সকৃত।
