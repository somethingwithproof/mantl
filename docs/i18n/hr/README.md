<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licencija: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Ponovljive postavke Kubernetes platforme sa sljedivim dokazima usklađenosti.

## Pregled

Mantl je prevoditelj platforme vođen specifikacijom i izvodno okruženje za usklađenost. YAML specifikacija `MantlCluster` prevodi se u ulazne varijable za Terraform, manifeste ArgoCD Application i manifeste izolacije imenskih prostora za korisnike. Zasebni Go operator usklađenosti usklađuje mapiranja kontrola, nalaze, procjene i verzionirane revizijske dokaze u S3 pohrani.

- `plan` radi izvan mreže za prevođenje, pohranu ili usporedbu artefakata platforme bez izmjena infrastrukture.
- `apply` provodi predprovjeru i pokreće generirane ili pregledane artefakte platforme.
- `status` prati generirane aplikacije i broj izvješća o Kyverno pravilima.
- `compliance` pregledava pokrivenost usklađenosti, nalaze, otklanjanje, povijest i arhive dokaza.

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

CLI alat `mantl` uključuje ugrađene prijevode za 48 jezičnih oznaka (20 Tier 1, 15 afričkih, 10 azijskih i 3 južnoslavenske) u paketu `pkg/i18n`. Odabir jezika poštuje varijable `LC_ALL`, `LC_MESSAGES` i `LANG` te trajnu zastavicu `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Razvoj i licencija

Pokrenite Go jedinične testove i provjere naredbama `make go-test`, `make go-coverage` i `golangci-lint run`. Pogledajte cjelovitu [englesku inačicu README](../../../README.md) i [kazalo dokumentacije](../../../docs/README.md). Objavljeno pod licencijom [Apache-2.0](../../../LICENSE).
