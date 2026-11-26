// SPDX-License-Identifier: Apache-2.0

package i18n

import (
	"encoding/json"
	"regexp"
	"sort"
	"testing"
)

var placeholderPattern = regexp.MustCompile(`%[0-9.]*[a-zA-Z]`)

func extractPlaceholders(s string) []string {
	matches := placeholderPattern.FindAllString(s, -1)
	sort.Strings(matches)
	return matches
}

func TestAllLocalesLoadWithParity(t *testing.T) {
	locales := SupportedLocales()
	if len(locales) != 20 {
		t.Fatalf("expected 20 supported locales, got %d", len(locales))
	}

	enRaw, err := localeFS.ReadFile("locales/en.json")
	if err != nil {
		t.Fatalf("read en.json: %v", err)
	}
	var enMap map[string]string
	if err := json.Unmarshal(enRaw, &enMap); err != nil {
		t.Fatalf("parse en.json: %v", err)
	}
	if len(enMap) == 0 {
		t.Fatal("en.json is empty")
	}

	for _, loc := range locales {
		t.Run(loc, func(t *testing.T) {
			raw, readErr := localeFS.ReadFile("locales/" + loc + ".json")
			if readErr != nil {
				t.Fatalf("read %s.json: %v", loc, readErr)
			}
			var locMap map[string]string
			if parseErr := json.Unmarshal(raw, &locMap); parseErr != nil {
				t.Fatalf("parse %s.json: %v", loc, parseErr)
			}
			if len(locMap) != len(enMap) {
				t.Fatalf("%s key count = %d, want %d", loc, len(locMap), len(enMap))
			}
			for key, enVal := range enMap {
				locVal, ok := locMap[key]
				if !ok {
					t.Errorf("%s missing key %q", loc, key)
					continue
				}
				if locVal == "" {
					t.Errorf("%s empty value for key %q", loc, key)
				}
				wantPH := extractPlaceholders(enVal)
				gotPH := extractPlaceholders(locVal)
				if len(wantPH) != len(gotPH) {
					t.Errorf("%s key %q placeholder mismatch: got %v, want %v", loc, key, gotPH, wantPH)
					continue
				}
				for i := range wantPH {
					if wantPH[i] != gotPH[i] {
						t.Errorf("%s key %q placeholder mismatch: got %v, want %v", loc, key, gotPH, wantPH)
						break
					}
				}
			}
			cat := New(loc)
			if cat.Locale() != loc {
				t.Errorf("New(%q).Locale() = %q", loc, cat.Locale())
			}
		})
	}
}

func TestResolveLocaleAndRTL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", "en"},
		{"C", "en"},
		{"POSIX", "en"},
		{"C.UTF-8", "en"},
		{"en_US.UTF-8", "en"},
		{"de_DE.UTF-8", "de"},
		{"pt_BR.UTF-8", "pt-BR"},
		{"zh_CN.UTF-8", "zh-CN"},
		{"zh_TW.UTF-8", "zh-TW"},
		{"zh-Hant", "zh-TW"},
		{"ar_EG.UTF-8", "ar"},
		{"hi_IN.UTF-8", "hi"},
		{"unknown", "en"},
	}
	for _, tc := range cases {
		if got := ResolveLocale(tc.input); got != tc.want {
			t.Errorf("ResolveLocale(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}

	for _, loc := range SupportedLocales() {
		wantRTL := loc == "ar"
		if IsRTL(loc) != wantRTL {
			t.Errorf("IsRTL(%q) = %v, want %v", loc, IsRTL(loc), wantRTL)
		}
	}
}

func TestDetectLocalePrecedence(t *testing.T) {
	t.Setenv("LC_ALL", "")
	t.Setenv("LC_MESSAGES", "")
	t.Setenv("LANG", "ko_KR.UTF-8")
	if got := DetectLocale(); got != "ko" {
		t.Fatalf("LANG only = %q, want ko", got)
	}

	t.Setenv("LC_MESSAGES", "pl_PL.UTF-8")
	if got := DetectLocale(); got != "pl" {
		t.Fatalf("LC_MESSAGES override = %q, want pl", got)
	}

	t.Setenv("LC_ALL", "POSIX")
	if got := DetectLocale(); got != "en" {
		t.Fatalf("LC_ALL=POSIX override = %q, want en", got)
	}
}

func TestEnglishBackwardCompatibility(t *testing.T) {
	cat := New("en")
	if got := cat.T("root.short"); got != "Mantl is a spec-driven platform compiler and runtime" {
		t.Fatalf("root.short = %q", got)
	}
	if got := cat.T("nonexistent.key"); got != "nonexistent.key" {
		t.Fatalf("unknown key = %q", got)
	}
}
