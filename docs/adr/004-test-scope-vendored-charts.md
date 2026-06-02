# ADR 004: Go Test Scope Excludes Vendored Charts

## Status
Accepted

## Context
The `platform/` tree vendors third-party Helm charts (Loki, Falco). Some carry
their own Go tests. `go test ./...` ran those tests, and they failed, turning CI
red for code Mantl does not own.

## Decision
The canonical test target excludes the vendored `platform/` packages while
keeping the project's own `apis/platform/` package. It is defined in the Makefile
as `GO_PKGS := go list ./... | grep -v 'github.com/thomasvincent/mantl/platform/'`
and run via `make go-test`. CI calls `make go-test`, not `go test ./...`.

## Rationale
Vendored chart tests are upstream concerns. Mantl's CI should gate Mantl's code.
Filtering at the test target is less invasive than dropping `go.mod` files into
each vendored chart, which chart updates would overwrite.

## Consequences
Contributors must use `make go-test` rather than `go test ./...`. A future
restructure that relocates vendored charts out of the module would let this
filter be removed.
