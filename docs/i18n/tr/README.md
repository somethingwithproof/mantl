<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lisans: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

İzlenebilir uyumluluk kanıtlarıyla tekrarlanabilir Kubernetes platform yapılandırması.

## Genel bakış

Mantl, belirtim odaklı bir platform derleyicisi ve uyumluluk çalışma zamanıdır. Bir `MantlCluster` YAML belirtimi; Terraform değişken girdilerine, ArgoCD Application bildirimlerine ve kiracı ad alanı (namespace) yalıtım bildirimlerine derlenir. Ayrı bir Go uyumluluk operatörü ise kontrol eşlemelerini, bulguları, değerlendirmeleri ve sürümlenmiş S3 denetim kanıtlarını uzlaştırır.

- `plan`, altyapıyı değiştirmeden platform yapılarını derlemek, kaydetmek veya karşılaştırmak için çevrimdışı çalışır.
- `apply`, ön kontrolleri çalıştırır ve üretilen veya gözden geçirilen platform yapılarını başlatır.
- `status`, üretilen uygulamaları ve Kyverno ilke raporu sayılarını gözlemler.
- `compliance`, uyumluluk kapsamını, bulguları, iyileştirmeleri, geçmişi ve kanıt arşivlerini inceler.

## Hızlı başlangıç

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Örnek `platform.yaml`:

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

## Yerelleştirme

`mantl` CLI aracı, `pkg/i18n` içinde 45 yerel ayar (20 Tier 1, 15 Afrika ve 10 Asya yerel ayarı) için çevirileri gömülü olarak içerir. Yerel ayar seçimi `LC_ALL`, `LC_MESSAGES`, `LANG` değişkenlerini ve kalıcı `--lang` bayrağını dikkate alır:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Geliştirme ve Lisans

Go birim testlerini ve kontrollerini `make go-test`, `make go-coverage` ve `golangci-lint run` ile çalıştırın. Tam [İngilizce README](../../../README.md) ve [belgelendirme dizini](../../../docs/README.md) dosyalarına bakın. [Apache-2.0](../../../LICENSE) altında lisanslanmıştır.
