// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"
)

func TestApplyLocalizationAndEnglishDefault(t *testing.T) {
	t.Cleanup(func() {
		applyLocalization("en")
	})

	enCat := applyLocalization("en")
	if enCat.Locale() != "en" {
		t.Fatalf("expected en locale, got %q", enCat.Locale())
	}
	if rootCmd.Short != "Mantl is a spec-driven platform compiler and runtime" {
		t.Fatalf("unexpected English rootCmd.Short: %q", rootCmd.Short)
	}

	deCat := applyLocalization("de")
	if deCat.Locale() != "de" {
		t.Fatalf("expected de locale, got %q", deCat.Locale())
	}
	if rootCmd.Short == "Mantl is a spec-driven platform compiler and runtime" {
		t.Fatal("expected localized German rootCmd.Short")
	}

	arCat := applyLocalization("ar")
	if !arCat.IsRTL() {
		t.Fatal("expected Arabic catalog to report RTL")
	}
}

func TestExtractLangArg(t *testing.T) {
	if got := extractLangArg([]string{"--lang", "ja", "plan"}); got != "ja" {
		t.Fatalf("expected ja, got %q", got)
	}
	if got := extractLangArg([]string{"plan", "--lang=pt-BR"}); got != "pt-BR" {
		t.Fatalf("expected pt-BR, got %q", got)
	}
	if got := extractLangArg([]string{"plan"}); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
