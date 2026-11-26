<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licenca: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Ponovljive postavke Kubernetes platforme uz sljedive dokaze usklađenosti.

## Pregled

Mantl je kompajler platforme zasnovan na specifikaciji i izvršno okruženje za usklađenost. YAML specifikacija `MantlCluster` kompajlira se u ulazne varijable za Terraform, manifeste ArgoCD Application i manifeste izolacije prostora imena za zakupce. Odvojeni Go operator za usklađenost usklađuje mapiranja kontrola, nalaze, ocjene i verzionisane revizijske dokaze u S3 skladištu.

- `plan` radi van mreže za kompajliranje, čuvanje ili poređenje artefakata platforme bez promjena u infrastrukturi.
- `apply` obavlja predprovjeru i pokreće generisane ili pregledane artefakte platforme.
- `status` prati generisane aplikacije i broj izvještaja o Kyverno pravilima.
- `compliance` pregleda pokrivenost usklađenosti, nalaze, sanaciju, historiju i arhive dokaza.

## Brzi početak

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Primjer datoteke `platform.yaml`:

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

## Lokalizacija

CLI alat `mantl` sadrži ugrađene prijevode za 48 jezičkih oznaka (20 Tier 1, 15 afričkih, 10 azijskih i 3 južnoslavenske) u paketu `pkg/i18n`. Izbor jezika poštuje varijable `LC_ALL`, `LC_MESSAGES` i `LANG` te trajnu zastavicu `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Razvoj i licenca

Pokrenite Go jedinične testove i provjere pomoću `make go-test`, `make go-coverage` i `golangci-lint run`. Pogledajte punu [englesku verziju README](../../../README.md) i [indeks dokumentacije](../../../docs/README.md). Objavljeno pod licencom [Apache-2.0](../../../LICENSE).
