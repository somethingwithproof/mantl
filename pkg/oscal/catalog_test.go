// SPDX-License-Identifier: Apache-2.0

package oscal

import (
	"encoding/json"
	"github.com/thomasvincent/mantl/pkg/framework"
	"testing"
	"time"
)

func TestRoundTripDoesNotInventMappings(t *testing.T) {
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Name: "SOC2", Version: "2017", Controls: []framework.Control{{ID: "CC6.1", Name: "Access", Mappings: []framework.Mapping{{Type: "policy"}}}}}, ContentDigest: "sha256:content"}
	document, err := Export(fw, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := Import(data)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Controls[0].ID != "CC6.1" || len(draft.Controls[0].Mappings) != 0 {
		t.Fatal("catalog import invented enforcement")
	}
	document.Catalog.Controls = append(document.Catalog.Controls, document.Catalog.Controls[0])
	data, err = json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Import(data); err == nil {
		t.Fatal("duplicate control accepted")
	}
}
