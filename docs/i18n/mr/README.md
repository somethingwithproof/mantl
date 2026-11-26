<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![परवाना: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

शोधता येण्याजोग्या अनुपालन पुराव्यांसह पुनरावृत्ती करण्यायोग्य Kubernetes प्लॅटफॉर्म कॉन्फिगरेशन.

## आढावा

Mantl हा विनिर्देश-आधारित प्लॅटफॉर्म कंपायलर आणि अनुपालन रनटाइम आहे. `MantlCluster` YAML विनिर्देशाचे संकलन Terraform व्हेरिएबल इनपुट्स, ArgoCD Application मॅनिफेस्ट्स आणि टेनंट नेमस्पेस विलगीकरण मॅनिफेस्ट्समध्ये होते. एक स्वतंत्र Go अनुपालन ऑपरेटर नियंत्रण मॅपिंग्ज, निष्कर्ष, मूल्यांकने आणि आवृत्तीबद्ध S3 ऑडिट पुराव्यांचा ताळमेळ राखतो.

- `plan` पायाभूत सुविधांमध्ये कोणताही बदल न करता प्लॅटफॉर्म आर्टिफॅक्ट्स संकलित करण्यासाठी, जतन करण्यासाठी किंवा त्यांची तुलना करण्यासाठी ऑफलाइन चालते.
- `apply` प्रीफ्लाइट तपासण्या चालवते आणि तयार केलेले किंवा तपासलेले प्लॅटफॉर्म आर्टिफॅक्ट्स बूटस्ट्रॅप करते.
- `status` तयार केलेल्या अ‍ॅप्लिकेशन्सचे आणि Kyverno धोरण अहवाल संख्येचे निरीक्षण करते.
- `compliance` अनुपालन व्याप्ती, निष्कर्ष, निवारण, इतिहास आणि पुरावा संग्रहणांची तपासणी करते.

## जलद सुरुवात

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

उदाहरण `platform.yaml`:

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

## स्थानिकीकरण

`mantl` CLI मध्ये `pkg/i18n` अंतर्गत 45 लोकॅल्सचे (20 टियर 1, 15 आफ्रिकन आणि 10 आशियाई लोकॅल्स) भाषांतर समाविष्ट आहे. लोकॅल निवड `LC_ALL`, `LC_MESSAGES`, `LANG` आणि स्थायी `--lang` फ्लॅगचे पालन करते:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## विकास आणि परवाना

`make go-test`, `make go-coverage` आणि `golangci-lint run` द्वारे Go युनिट चाचण्या आणि तपासण्या चालवा. संपूर्ण [इंग्रजी README](../../../README.md) आणि [दस्तऐवजीकरण अनुक्रमणिका](../../../docs/README.md) पहा. [Apache-2.0](../../../LICENSE) अंतर्गत परवानाकृत.
