<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

கண்டறியக்கூடிய இணக்கச் சான்றுகளுடன் கூடிய மீளுருவாக்கத்தக்க Kubernetes தள உள்ளமைவு.

## கண்ணோட்டம்

Mantl என்பது விவரக்குறிப்பு அடிப்படையிலான தளத் தொகுப்பி மற்றும் இணக்க இயக்க நேரமாகும். ஒரு `MantlCluster` YAML விவரக்குறிப்பு Terraform மாறி உள்ளீடுகள், ArgoCD Application வெளிப்பாடுகள் மற்றும் குத்தகைதாரர் பெயர்வெளித் தனிமைப்படுத்தல் வெளிப்பாடுகளாகத் தொகுக்கப்படுகிறது. ஒரு தனி Go இணக்க இயக்குநர் கட்டுப்பாட்டு வரைபடங்கள், கண்டறிதல்கள், மதிப்பீடுகள் மற்றும் பதிப்பாக்கப்பட்ட S3 தணிக்கைச் சான்றுகளைச் சீரமைக்கிறது.

- உள்கட்டமைப்பை மாற்றாமல் தளக் கூறுகளைத் தொகுக்க, சேமிக்க அல்லது ஒப்பிட `plan` இணையத் தொடர்பின்றி இயங்குகிறது.
- `apply` முன்னோட்டச் சரிபார்ப்புகளை இயக்கி, உருவாக்கப்பட்ட அல்லது மதிப்பாய்வு செய்யப்பட்ட தளக் கூறுகளைத் தொடங்குகிறது.
- `status` உருவாக்கப்பட்ட பயன்பாடுகள் மற்றும் Kyverno கொள்கை அறிக்கை எண்ணிக்கைகளைக் கண்காணிக்கிறது.
- `compliance` இணக்க உள்ளடக்கம், கண்டறிதல்கள், தீர்வு நடவடிக்கை, வரலாறு மற்றும் சான்றுக் காப்பகங்களை ஆய்வு செய்கிறது.

## விரைவுத் தொடக்கம்

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

எடுத்துக்காட்டு `platform.yaml`:

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

## உள்ளூர்மயமாக்கல்

`mantl` CLI ஆனது `pkg/i18n` இல் 45 மொழியிடங்களுக்கான (20 Tier 1, 15 ஆப்பிரிக்க மற்றும் 10 ஆசிய மொழியிடங்கள்) மொழிபெயர்ப்புகளை உட்பொதித்துள்ளது. மொழியிடத் தேர்வு `LC_ALL`, `LC_MESSAGES`, `LANG` மற்றும் நிலையான `--lang` கொடியைப் பின்பற்றுகிறது:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## மேம்பாடு மற்றும் உரிமம்

`make go-test`, `make go-coverage` மற்றும் `golangci-lint run` மூலம் Go அலகு சோதனைகளையும் சரிபார்ப்புகளையும் இயக்கவும். முழுமையான [ஆங்கில README](../../../README.md) மற்றும் [ஆவண அட்டவணை](../../../docs/README.md) ஆகியவற்றைப் பார்க்கவும். [Apache-2.0](../../../LICENSE) உரிமத்தின் கீழ் வெளியிடப்பட்டது.
