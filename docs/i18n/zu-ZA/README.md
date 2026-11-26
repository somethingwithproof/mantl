<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Ukulungiselelwa kwenkundla ye-Kubernetes okungaphindaphindwa nobufakazi bokuhambisana nemithetho obulandelekayo.

## Uhlolojikelele

I-Mantl iyisihlanganisi senkundla esiqhutshwa yizicaciso kanye nesikhathi sokusebenza sokuhambisana nemithetho. Isicaciso se-YAML se-`MantlCluster` sihlanganiswa sibe okokufaka okuguquguqukayo kwe-Terraform, imibhalo ye-ArgoCD Application, kanye nemibhalo yokuhlukaniswa kwe-namespace yabaqashi. Umsebenzisi wokuhambisana nemithetho we-Go ohlukile uhlanganisa ukumakwa kwezilawuli, okutholakele, ukuhlola, kanye nobufakazi bokucwaningwa kwamabhuku be-S3 obunezinguqulo.

- `plan` isebenza ngaphandle kwe-inthanethi ukuhlanganisa, ukulondoloza, noma ukuqhathanisa izinto zenkundla ngaphandle kokushintsha ingqalasizinda.
- `apply` yenza ukuhlola kwangaphambi kokuqalisa futhi iqale izinto zenkundla ezikhiqiziwe noma ezibuyekeziwe.
- `status` ibheka izinhlelo zokusebenza ezikhiqiziwe kanye nezibalo zemibiko yenqubomgomo ye-Kyverno.
- `compliance` ihlola ukumbozwa kokuhambisana nemithetho, okutholakele, ukulungiswa, umlando, kanye nezinqolobane zobufakazi.

## Ukuqala ngokushesha

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Isibonelo se-`platform.yaml`:

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

## Ukwenziwa ngokwendawo

I-CLI ye-`mantl` ifaka ukuhunyushwa kwezilimi zendawo ezingu-45 (ezingu-20 ze-Tier 1, ezingu-15 zase-Afrika, kanye nezingu-10 zase-Asia) ku-`pkg/i18n`. Ukukhethwa kolimi lwendawo kuhlonipha i-`LC_ALL`, i-`LC_MESSAGES`, i-`LANG`, kanye nefulegi eqhubekayo ye-`--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Ukuthuthukiswa kanye Nelayisense

Qalisa ukuhlolwa kwe-Go unit kanye nokuhlolwa nge-`make go-test`, i-`make go-coverage`, kanye ne-`golangci-lint run`. Bheka [i-README yesiNgisi](../../../README.md) ephelele kanye ne-[uhlu lwemibhalo](../../../docs/README.md). Inikezwe ilayisense ngaphansi kwe-[Apache-2.0](../../../LICENSE).
