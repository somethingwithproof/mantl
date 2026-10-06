// SPDX-License-Identifier: Apache-2.0

package evidence

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"
)

type exportStore map[string][]byte

func (s exportStore) Put(context.Context, string, []byte, time.Time) (ObjectRef, error) {
	return ObjectRef{}, fmt.Errorf("unused")
}
func (s exportStore) Get(_ context.Context, ref ObjectRef) ([]byte, error) {
	data, ok := s[ref.URI]
	if !ok {
		return nil, fmt.Errorf("missing object")
	}
	return data, nil
}
func TestExportRejectsTamperedObjects(t *testing.T) {
	payload := []byte(`{"items":[]}`)
	object := ObjectRef{URI: "object", Hash: Hash(payload), Version: "v1"}
	manifest, _ := json.Marshal(Manifest{SchemaVersion: 1, Objects: []ObjectRef{object}})
	ref := ObjectRef{URI: "manifest", Hash: Hash(manifest), Version: "v1"}
	store := exportStore{"manifest": manifest, "object": payload}
	var out bytes.Buffer
	if err := Export(context.Background(), store, ref, &out); err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(&out)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	files := 0
	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		files++
	}
	if files != 2 {
		t.Fatal("archive missing evidence")
	}
	store["object"] = []byte("tampered")
	if err := Export(context.Background(), store, ref, io.Discard); err == nil {
		t.Fatal("tampered object exported")
	}
	store["manifest"] = []byte("tampered")
	if err := Export(context.Background(), store, ref, io.Discard); err == nil {
		t.Fatal("tampered manifest exported")
	}
}
