<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

추적 가능한 규정 준수 증적을 갖춘 재현 가능한 Kubernetes 플랫폼 구성 도구입니다.

## 개요

Mantl은 사양 기반 플랫폼 컴파일러이자 규정 준수 런타임입니다. `MantlCluster` YAML 사양은 Terraform 변수 입력, ArgoCD Application 매니페스트 및 테넌트 네임스페이스 격리 매니페스트로 컴파일됩니다. 별도의 Go 규정 준수 오퍼레이터가 제어 매핑, 결과(findings), 평가 및 버전 관리되는 S3 감사 증적을 조정합니다.

- `plan`은 인프라를 변경하지 않고 오프라인으로 실행되어 플랫폼 아티팩트를 컴파일, 저장 또는 비교합니다.
- `apply`는 사전 점검(preflight check)을 실행하고 생성되거나 검토된 플랫폼 아티팩트를 부트스트랩합니다.
- `status`는 생성된 애플리케이션과 Kyverno 정책 보고서 수를 관찰합니다.
- `compliance`는 규정 준수 적용 범위, 결과(findings), 개선 조치, 이력 및 증적 아카이브를 검사합니다.

## 빠른 시작

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

`platform.yaml` 예시:

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

## 현지화

`mantl` CLI는 `pkg/i18n`에 45개 로캘(Tier 1 20개, 아프리카 15개, 아시아 10개)의 번역을 내장하고 있습니다. 로캘 선택은 `LC_ALL`, `LC_MESSAGES`, `LANG` 및 영구 플래그 `--lang`을 따릅니다:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## 개발 및 라이선스

`make go-test`, `make go-coverage`, `golangci-lint run`으로 Go 단위 테스트와 검사를 실행합니다. 자세한 내용은 전체 [영문 README](../../../README.md)와 [문서 색인](../../../docs/README.md)을 참조하세요. [Apache-2.0](../../../LICENSE)에 따라 라이선스가 부여됩니다.
