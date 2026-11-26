<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Shatiga: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Qaabeynta madasha Kubernetes ee lagu celcelin karo iyadoo leh caddayn u hoggaansanaan oo la raadraaci karo.

## Dulmar

Mantl waa isku-duwe madal ku dhisan qeexitaan iyo nidaam fulin oo u hoggaansanaan. Qeexitaanka YAML ee `MantlCluster` wuxuu isu beddelaa gelinta doorsoomayaasha Terraform, muuqaallada ArgoCD Application, iyo muuqaallada go'doominta namespace-ka kiraystayaasha. Hawlwadeen u hoggaansanaan oo Go ku qoran oo gaar ah ayaa mideeya jaangooyooyinka xakamaynta, natiijooyinka baaritaanka, qiimeynta, iyo caddaynta hantidhawrka S3 ee noocyeysan.

- `plan` wuxuu u shaqeeyaa si offline ah si uu u soo ururiyo, u kaydiyo, ama u isbarbardhigo agabka madasha isagoon beddelin kaabayaasha.
- `apply` wuxuu sameeyaa hubinta horudhaca ah wuxuuna bilaabaa agabka madasha ee la soo saaray ama dib loo eegay.
- `status` wuxuu kormeeraa codsiyada la soo saaray iyo tirada warbixinnada siyaasadda Kyverno.
- `compliance` wuxuu baaraa daboolista u hoggaansanaanta, natiijooyinka, sixitaanka, taariikhda, iyo kaydka caddaynta.

## Bilow degdeg ah

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Tusaale `platform.yaml`:

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

## Deegaameynta

CLI-ga `mantl` wuxuu ku dhex kaydiyaa tarjumaado loogu talagalay 45 deegaan (20 Heerka 1-aad, 15 Afrikaan, iyo 10 Aasiyaan ah) gudaha `pkg/i18n`. Doorashada deegaanku waxay tixgelisaa `LC_ALL`, `LC_MESSAGES`, `LANG`, iyo calanka joogtada ah ee `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Horumarinta iyo Shatiga

Ku socodsii imtixaannada cutubka Go iyo hubinta adigoo adeegsanaya `make go-test`, `make go-coverage`, iyo `golangci-lint run`. Eeg [README-ga Ingiriisiga](../../../README.md) oo dhammaystiran iyo [tusmada dukumiintiyada](../../../docs/README.md). Waxaa lagu shatiyeeyay [Apache-2.0](../../../LICENSE).
