package evidence

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

var hashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

// Export verifies the manifest and every object before adding it to the archive.
func Export(ctx context.Context, store Store, ref ObjectRef, out io.Writer) error {
	data, err := store.Get(ctx, ref)
	if err != nil {
		return err
	}
	if Hash(data) != ref.Hash {
		return fmt.Errorf("manifest hash mismatch")
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("decode manifest: %w", err)
	}
	if manifest.SchemaVersion != 1 || len(manifest.Objects) > 10000 {
		return fmt.Errorf("unsupported evidence manifest")
	}
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	write := func(name string, payload []byte) error {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(payload))}); err != nil {
			return err
		}
		_, err := tw.Write(payload)
		return err
	}
	if err := write("manifest.json", data); err != nil {
		return err
	}
	seen := map[string]bool{}
	total := len(data)
	for _, obj := range manifest.Objects {
		if !hashPattern.MatchString(obj.Hash) || obj.Version == "" {
			return fmt.Errorf("invalid evidence object identity")
		}
		payload, err := store.Get(ctx, obj)
		if err != nil {
			return err
		}
		if Hash(payload) != obj.Hash {
			return fmt.Errorf("object hash mismatch")
		}
		total += len(payload)
		if total > 512*1024*1024 {
			return fmt.Errorf("evidence export exceeds size limit")
		}
		if !seen[obj.Hash] {
			if err := write("objects/"+obj.Hash+".json", payload); err != nil {
				return err
			}
			seen[obj.Hash] = true
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}
