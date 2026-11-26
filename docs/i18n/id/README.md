<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lisensi: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Konfigurasi platform Kubernetes yang dapat direproduksi dengan bukti kepatuhan yang dapat ditelusuri.

## Ikhtisar

Mantl adalah kompilator platform berbasis spesifikasi dan runtime kepatuhan. Spesifikasi YAML `MantlCluster` dikompilasi menjadi input variabel Terraform, manifes ArgoCD Application, dan manifes isolasi namespace tenant. Operator kepatuhan Go yang terpisah merekonsiliasi pemetaan kontrol, temuan, evaluasi, dan bukti audit S3 berversi.

- `plan` berjalan secara luring untuk mengompilasi, menyimpan, atau membandingkan artefak platform tanpa mengubah infrastruktur.
- `apply` menjalankan pemeriksaan awal dan melakukan bootstrap pada artefak platform yang dihasilkan atau telah ditinjau.
- `status` mengamati aplikasi yang dihasilkan dan jumlah laporan kebijakan Kyverno.
- `compliance` memeriksa cakupan kepatuhan, temuan, remediasi, riwayat, dan arsip bukti.

## Mulai cepat

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Contoh `platform.yaml`:

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

## Pelokalan

CLI `mantl` menyertakan terjemahan untuk 45 lokal (20 lokal Tier 1, 15 lokal Afrika, dan 10 lokal Asia) di `pkg/i18n`. Pemilihan lokal mengikuti `LC_ALL`, `LC_MESSAGES`, `LANG`, dan flag persisten `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Pengembangan dan Lisensi

Jalankan pengujian unit dan pemeriksaan Go melalui `make go-test`, `make go-coverage`, dan `golangci-lint run`. Lihat [README bahasa Inggris](../../../README.md) lengkap dan [indeks dokumentasi](../../../docs/README.md). Dilisensikan di bawah [Apache-2.0](../../../LICENSE).
