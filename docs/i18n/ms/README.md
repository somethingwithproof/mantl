<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lesen: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Konfigurasi platform Kubernetes yang boleh diulang dengan bukti pematuhan yang boleh dijejak.

## Gambaran Keseluruhan

Mantl ialah pengkompil platform berasaskan spesifikasi dan masa jalan pematuhan. Spesifikasi YAML `MantlCluster` dikompilkan menjadi input pemboleh ubah Terraform, manifes Aplikasi ArgoCD, dan manifes pengasingan ruang nama penyewa. Pengendali pematuhan Go yang berasingan menyelaras pemetaan kawalan, penemuan, penilaian, dan bukti audit S3 berversi.

- `plan` dijalankan secara luar talian untuk mengkompil, menyimpan, atau membandingkan artifak platform tanpa mengubah infrastruktur.
- `apply` menjalankan semakan prapenerbangan dan membutstrap artifak platform yang dijana atau disemak.
- `status` memerhati aplikasi yang dijana dan kiraan laporan dasar Kyverno.
- `compliance` memeriksa liputan pematuhan, penemuan, pemulihan, sejarah, dan arkib bukti.

## Mula Pantas

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

## Penyetempatan

CLI `mantl` membenamkan terjemahan untuk 45 lokal (20 lokal Peringkat 1, 15 lokal Afrika, dan 10 lokal Asia) dalam `pkg/i18n`. Pemilihan lokal mematuhi `LC_ALL`, `LC_MESSAGES`, `LANG`, dan bendera `--lang` yang berterusan:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Pembangunan dan Lesen

Jalankan ujian unit dan semakan Go melalui `make go-test`, `make go-coverage`, dan `golangci-lint run`. Lihat [README Bahasa Inggeris](../../../README.md) penuh dan [indeks dokumentasi](../../../docs/README.md). Dilesenkan di bawah [Apache-2.0](../../../LICENSE).
