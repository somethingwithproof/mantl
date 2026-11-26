<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Lasisi: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Saitin dandamalin Kubernetes mai maimaitawa tare da shaidar bin ƙa'ida mai bin sawu.

## Taƙaitaccen Bayani

Mantl mai haɗa dandamali ne mai dogaro da ƙayyadaddun bayanai da kuma yanayin gudanar da bin ƙa'ida. Ƙayyadadden bayanin YAML na `MantlCluster` yana haɗuwa zuwa shigarwar masu canji na Terraform, takardun bayanin ArgoCD Application, da takardun keɓancewar namespace na masu haya. Wani daban na Go compliance operator yana daidaita taswirorin kulawa, bincike, kimantawa, da shaidar binciken S3 mai siga.

- `plan` yana aiki ba tare da intanet ba don haɗawa, adanawa, ko kwatanta kayayyakin dandamali ba tare da canza ababen more rayuwa ba.
- `apply` yana gudanar da binciken farko kuma yana ƙaddamar da kayayyakin dandamalin da aka ƙirƙira ko aka duba.
- `status` yana lura da manhajojin da aka ƙirƙira da adadin rahotannin manufofin Kyverno.
- `compliance` yana bincika iyakar bin ƙa'ida, bincike, gyare-gyare, tarihi, da rumbun adana shaida.

## Fara Amfani da Sauri

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Misalin `platform.yaml`:

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

## Daidaita Harshe

`mantl` CLI yana ɗauke da fassarori don harsuna 45 (20 na Mataki na 1, 15 na Afirka, da 10 na Asiya) a cikin `pkg/i18n`. Zaɓin harshe yana mutunta `LC_ALL`, `LC_MESSAGES`, `LANG`, da tutar `--lang` mai dorewa:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Haɓakawa da Lasisi

Gudanar da gwaje-gwajen Go da bincike ta hanyar `make go-test`, `make go-coverage`, da `golangci-lint run`. Duba cikakken [README na Turanci](../../../README.md) da [jerin takardun bayani](../../../docs/README.md). An ba da lasisi ƙarƙashin [Apache-2.0](../../../LICENSE).
