<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

ट्रेस करने योग्य अनुपालन साक्ष्य के साथ दोहराए जाने योग्य Kubernetes प्लेटफ़ॉर्म कॉन्फ़िगरेशन।

## अवलोकन

Mantl एक विनिर्देश-संचालित प्लेटफ़ॉर्म कंपाइलर और अनुपालन रनटाइम है। एक `MantlCluster` YAML विनिर्देश Terraform वेरिएबल इनपुट, ArgoCD Application मैनिफ़ेस्ट और टेनेंट नेमस्पेस आइसोलेशन मैनिफ़ेस्ट में संकलित होता है। एक अलग Go अनुपालन ऑपरेटर नियंत्रण मैपिंग, निष्कर्षों, मूल्यांकनों और संस्करणित S3 ऑडिट साक्ष्यों का मिलान करता है।

- `plan` अवसंरचना में बदलाव किए बिना प्लेटफ़ॉर्म आर्टिफैक्ट्स को संकलित करने, सहेजने या तुलना करने के लिए ऑफ़लाइन चलता है।
- `apply` प्रीफ़्लाइट जाँच चलाता है और उत्पन्न या समीक्षित प्लेटफ़ॉर्म आर्टिफैक्ट्स को बूटस्ट्रैप करता है।
- `status` उत्पन्न अनुप्रयोगों और Kyverno नीति रिपोर्ट गणनाओं का निरीक्षण करता है।
- `compliance` अनुपालन कवरेज, निष्कर्षों, समाधान, इतिहास और साक्ष्य अभिलेखागार का निरीक्षण करता है।

## त्वरित आरंभ

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

## स्थानीयकरण

`mantl` CLI `pkg/i18n` में 45 लोकेल (20 टियर 1, 15 अफ़्रीकी और 10 एशियाई लोकेल) के लिए अनुवाद एम्बेड करता है। लोकेल चयन `LC_ALL`, `LC_MESSAGES`, `LANG` और स्थायी `--lang` फ़्लैग का सम्मान करता है:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## विकास और लाइसेंस

`make go-test`, `make go-coverage` और `golangci-lint run` के माध्यम से Go यूनिट परीक्षण और जाँच चलाएँ। पूर्ण [अंग्रेज़ी README](../../../README.md) और [दस्तावेज़ीकरण सूचकांक](../../../docs/README.md) देखें। [Apache-2.0](../../../LICENSE) के अंतर्गत लाइसेंस प्राप्त।
