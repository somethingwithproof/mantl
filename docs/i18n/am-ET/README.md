<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![ፈቃድ: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

ሊደገም የሚችል የKubernetes መድረክ ውቅር ከተከታታይ የተገዢነት ማስረጃ ጋር።

## አጠቃላይ እይታ

Mantl በዝርዝር መግለጫ ላይ የተመሠረተ የመድረክ አቀናባሪ እና የተገዢነት አሂድ ነው። የ`MantlCluster` YAML ዝርዝር መግለጫ ወደ Terraform ተለዋዋጭ ግብዓቶች፣ የArgoCD Application ማኒፌስቶች እና የተከራይ የስም ቦታ መነጠያ ማኒፌስቶች ይቀናበራል። የተለየ የGo ተገዢነት ኦፕሬተር የቁጥጥር ካርታዎችን፣ ግኝቶችን፣ ግምገማዎችን እና ስሪት ያላቸውን የS3 ኦዲት ማስረጃዎችን ያስታርቃል።

- `plan` መሠረተ ልማቱን ሳይቀይር የመድረክ ውጤቶችን ለማቀናበር፣ ለማስቀመጥ ወይም ለማነጻጸር ከመስመር ውጭ ይሠራል።
- `apply` የቅድመ-ፍተሻ ምርመራዎችን ያካሂዳል እና የተመነጩ ወይም የተገመገሙ የመድረክ ውጤቶችን ያስጀምራል።
- `status` የተመነጩ መተግበሪያዎችን እና የKyverno ፖሊሲ ሪፖርት ቆጠራዎችን ይመለከታል።
- `compliance` የተገዢነት ሽፋንን፣ ግኝቶችን፣ እርማቶችን፣ ታሪክን እና የማስረጃ ማህደሮችን ይመረምራል።

## ፈጣን ጅምር

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

የ`platform.yaml` ምሳሌ፦

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

## አካባቢያዊ ትርጉም

የ`mantl` CLI በ`pkg/i18n` ውስጥ ለ45 አካባቢዎች (20 ደረጃ 1፣ 15 የአፍሪካ እና 10 የእስያ አካባቢዎች) ትርጉሞችን ያካትታል። የአካባቢ ምርጫ `LC_ALL`፣ `LC_MESSAGES`፣ `LANG` እና ቋሚውን `--lang` አማራጭ ያከብራል፦

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## ልማት እና ፈቃድ

የGo አሃድ ሙከራዎችን እና ፍተሻዎችን በ`make go-test`፣ `make go-coverage` እና `golangci-lint run` ያስኪዱ። ሙሉውን [የእንግሊዝኛ README](../../../README.md) እና [የሰነድ ማውጫ](../../../docs/README.md) ይመልከቱ። በ[Apache-2.0](../../../LICENSE) ፈቃድ የተሰጠ።
