<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Cấu hình nền tảng Kubernetes có thể tái lập cùng bằng chứng tuân thủ có khả năng truy vết.

## Tổng quan

Mantl là trình biên dịch nền tảng và môi trường thực thi tuân thủ định hướng theo đặc tả. Một bản đặc tả YAML `MantlCluster` được biên dịch thành các biến đầu vào Terraform, các manifest ArgoCD Application và các manifest cô lập namespace theo tenant. Một operator tuân thủ riêng biệt viết bằng Go đảm nhiệm việc đối soát ánh xạ kiểm soát, phát hiện, đánh giá và bằng chứng kiểm toán S3 có quản lý phiên bản.

- `plan` chạy ngoại tuyến để biên dịch, lưu hoặc so sánh các artifact của nền tảng mà không làm thay đổi hạ tầng.
- `apply` chạy các bước kiểm tra tiền điều kiện và khởi tạo các artifact nền tảng đã được tạo hoặc đã được xem xét.
- `status` theo dõi các ứng dụng được tạo và số lượng báo cáo chính sách Kyverno.
- `compliance` kiểm tra mức độ bao phủ tuân thủ, các phát hiện, khắc phục, lịch sử và kho lưu trữ bằng chứng.

## Bắt đầu nhanh

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Ví dụ tệp `platform.yaml`:

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

## Bản địa hóa

CLI `mantl` nhúng sẵn bản dịch cho 45 ngôn ngữ địa phương (20 ngôn ngữ Bậc 1, 15 ngôn ngữ châu Phi và 10 ngôn ngữ châu Á) trong `pkg/i18n`. Việc chọn ngôn ngữ tuân theo `LC_ALL`, `LC_MESSAGES`, `LANG` và cờ cố định `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Phát triển và Giấy phép

Chạy các bài kiểm thử đơn vị và kiểm tra Go thông qua `make go-test`, `make go-coverage` và `golangci-lint run`. Xem đầy đủ [README tiếng Anh](../../../README.md) và [mục lục tài liệu](../../../docs/README.md). Phát hành theo giấy phép [Apache-2.0](../../../LICENSE).
