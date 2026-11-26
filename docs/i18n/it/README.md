<!-- SPDX-License-Identifier: Apache-2.0 -->

> [English](../../../README.md)

# Mantl

[![Mantl CI](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/somethingwithproof/mantl/actions/workflows/ci.yml)
[![Licenza: Apache-2.0](https://img.shields.io/github/license/somethingwithproof/mantl)](../../../LICENSE)

Configurazione ripetibile di piattaforme Kubernetes con evidenze di conformità tracciabili.

## Panoramica

Mantl è un compilatore di piattaforma guidato da specifiche e un runtime per la conformità. Una specifica YAML `MantlCluster` viene compilata in input di variabili Terraform, manifest Application per ArgoCD e manifest di isolamento dei namespace per i tenant. Un operatore Go separato dedicato alla conformità riconcilia mappature dei controlli, rilevazioni, valutazioni e prove di audit versionate su S3.

- `plan` viene eseguito offline per compilare, salvare o confrontare gli artefatti della piattaforma senza modificare l'infrastruttura.
- `apply` esegue i controlli preliminari e avvia il bootstrap degli artefatti della piattaforma generati o revisionati.
- `status` osserva le applicazioni generate e i conteggi dei report delle policy Kyverno.
- `compliance` ispeziona la copertura di conformità, le rilevazioni, la rimediazione, la cronologia e gli archivi delle prove.

## Avvio rapido

```bash
go run ./cmd/mantl plan examples/mantl-spec.yaml
go run ./cmd/mantl --help
```

Esempio di `platform.yaml`:

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

## Localizzazione

La CLI `mantl` integra le traduzioni per 45 localizzazioni (20 Tier 1, 15 africane e 10 asiatiche) in `pkg/i18n`. La selezione della lingua rispetta `LC_ALL`, `LC_MESSAGES`, `LANG` e il flag persistente `--lang`:

```bash
mantl --lang de --help
LANG=ja_JP.UTF-8 mantl --help
```

## Sviluppo e licenza

Esegui i test unitari e i controlli Go tramite `make go-test`, `make go-coverage` e `golangci-lint run`. Consulta il [README completo in inglese](../../../README.md) e l'[indice della documentazione](../../../docs/README.md). Rilasciato sotto licenza [Apache-2.0](../../../LICENSE).
