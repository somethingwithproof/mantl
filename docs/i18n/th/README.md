<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

การตั้งค่าแพลตฟอร์ม Kubernetes ที่ทำซ้ำได้พร้อมหลักฐานการปฏิบัติตามข้อกำหนดที่ตรวจสอบย้อนกลับได้

## ภาพรวม

Mantl เป็นตัวคอมไพล์แพลตฟอร์มที่ขับเคลื่อนด้วยข้อกำหนดและรันไทม์ด้านการปฏิบัติตามข้อกำหนด ข้อกำหนด YAML ของ `MantlCluster` จะคอมไพล์เป็นอินพุตตัวแปร Terraform ไฟล์แมนิเฟสต์ของ ArgoCD Application และไฟล์แมนิเฟสต์สำหรับการแยกเนมสเปซของผู้เช่า นอกจากนี้ยังมีตัวดำเนินการด้านการปฏิบัติตามข้อกำหนดที่เขียนด้วย Go แยกต่างหากคอยปรับสมดุลการจับคู่การควบคุม ผลการตรวจพบ การประเมินผล และหลักฐานการตรวจสอบบน S3 แบบมีเวอร์ชัน

- `plan` ทำงานแบบออฟไลน์เพื่อคอมไพล์ บันทึก หรือเปรียบเทียบอาร์ติแฟกต์ของแพลตฟอร์มโดยไม่เปลี่ยนแปลงโครงสร้างพื้นฐาน
- `apply` รันการตรวจสอบเบื้องต้นและตั้งค่าเริ่มต้นอาร์ติแฟกต์ของแพลตฟอร์มที่สร้างขึ้นหรือผ่านการตรวจสอบแล้ว
- `status` เฝ้าสังเกตแอปพลิเคชันที่สร้างขึ้นและจำนวนรายงานนโยบายของ Kyverno
- `compliance` ตรวจสอบความครอบคลุมของการปฏิบัติตามข้อกำหนด ผลการตรวจพบ การแก้ไขปัญหา ประวัติ และคลังจัดเก็บหลักฐาน

## เริ่มต้นใช้งานอย่างรวดเร็ว

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

ตัวอย่าง `platform.yaml`:

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

## การรองรับภาษาท้องถิ่น

CLI ของ `mantl` ฝังคำแปลสำหรับ 45 ภาษาท้องถิ่น (ภาษาท้องถิ่นระดับ Tier 1 จำนวน 20 ภาษา แอฟริกา 15 ภาษา และเอเชีย 10 ภาษา) ไว้ใน `pkg/i18n` การเลือกภาษาท้องถิ่นจะยึดตาม `LC_ALL`, `LC_MESSAGES`, `LANG` และแฟล็ก `--lang` แบบถาวร:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## การพัฒนาและสัญญาอนุญาต

รันการทดสอบยูนิตและการตรวจสอบของ Go ผ่าน `make go-test`, `make go-coverage` และ `golangci-lint run` ดูเพิ่มเติมได้ที่ [README ภาษาอังกฤษ](../../../README.md) ฉบับเต็มและ [ดัชนีเอกสารประกอบ](../../../docs/README.md) อนุญาตให้ใช้งานภายใต้สัญญาอนุญาต [Apache-2.0](../../../LICENSE)
