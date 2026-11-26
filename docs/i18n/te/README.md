<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

ట్రేస్ చేయగల సమ్మతి ఆధారాలతో కూడిన పునరావృత Kubernetes ప్లాట్‌ఫారమ్ కాన్ఫిగరేషన్.

## అవలోకనం

Mantl అనేది స్పెక్-ఆధారిత ప్లాట్‌ఫారమ్ కంపైలర్ మరియు సమ్మతి రన్‌టైమ్. `MantlCluster` YAML స్పెసిఫికేషన్ Terraform వేరియబుల్ ఇన్‌పుట్‌లు, ArgoCD Application మానిఫెస్ట్‌లు, మరియు టెనెంట్ నేమ్‌స్పేస్ ఐసోలేషన్ మానిఫెస్ట్‌లుగా కంపైల్ అవుతుంది. ఒక ప్రత్యేక Go సమ్మతి ఆపరేటర్ కంట్రోల్ మ్యాపింగ్‌లు, గుర్తింపులు, మూల్యాంకనాలు, మరియు వెర్షన్ చేయబడిన S3 ఆడిట్ ఆధారాలను సమన్వయం చేస్తుంది.

- `plan` మౌలిక సదుపాయాలను మార్చకుండానే ప్లాట్‌ఫారమ్ ఆర్టిఫాక్ట్‌లను కంపైల్ చేయడానికి, సేవ్ చేయడానికి లేదా పోల్చడానికి ఆఫ్‌లైన్‌లో నడుస్తుంది.
- `apply` ప్రీఫ్లైట్ తనిఖీలను నడుపుతుంది మరియు రూపొందించబడిన లేదా సమీక్షించబడిన ప్లాట్‌ఫారమ్ ఆర్టిఫాక్ట్‌లను బూట్‌స్ట్రాప్ చేస్తుంది.
- `status` రూపొందించబడిన అప్లికేషన్‌లు మరియు Kyverno విధాన నివేదికల గణనలను గమనిస్తుంది.
- `compliance` సమ్మతి కవరేజ్, గుర్తింపులు, నివారణ చర్యలు, చరిత్ర, మరియు ఆధార ఆర్కైవ్‌లను పరిశీలిస్తుంది.

## త్వరిత ప్రారంభం

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

ఉదాహరణ `platform.yaml`:

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

## స్థానికీకరణ

`mantl` CLI 45 లోకేల్స్ కోసం అనువాదాలను (20 టియర్ 1, 15 ఆఫ్రికన్, మరియు 10 ఆసియా లోకేల్స్) `pkg/i18n` లో పొందుపరుస్తుంది. లోకేల్ ఎంపిక `LC_ALL`, `LC_MESSAGES`, `LANG`, మరియు నిరంతర `--lang` ఫ్లాగ్‌ను అనుసరిస్తుంది:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## అభివృద్ధి మరియు లైసెన్స్

`make go-test`, `make go-coverage`, మరియు `golangci-lint run` ద్వారా Go యూనిట్ పరీక్షలు మరియు తనిఖీలను అమలు చేయండి. పూర్తి [ఆంగ్ల README](../../../README.md) మరియు [డాక్యుమెంటేషన్ సూచిక](../../../docs/README.md) ను చూడండి. [Apache-2.0](../../../LICENSE) క్రింద లైసెన్స్ పొందింది.
