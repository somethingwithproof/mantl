// Package oscal translates control catalogs without inventing policy mappings.
package oscal

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/framework"
	"time"
)

type Metadata struct {
	Title        string    `json:"title"`
	LastModified time.Time `json:"last-modified"`
	Version      string    `json:"version"`
	OSCALVersion string    `json:"oscal-version"`
}
type Control struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type Catalog struct {
	UUID     string    `json:"uuid"`
	Metadata Metadata  `json:"metadata"`
	Controls []Control `json:"controls,omitempty"`
	Groups   []Group   `json:"groups,omitempty"`
}
type Group struct {
	ID       string    `json:"id,omitempty"`
	Title    string    `json:"title"`
	Controls []Control `json:"controls,omitempty"`
	Groups   []Group   `json:"groups,omitempty"`
}
type Document struct {
	Catalog Catalog `json:"catalog"`
}

func Export(fw *framework.Framework, at time.Time) (Document, error) {
	if fw == nil || fw.Version == "" || len(fw.Controls) == 0 {
		return Document{}, fmt.Errorf("empty control catalog")
	}
	// UUIDv5 namespace URL with the content digest as the stable catalog name.
	namespace, _ := hex.DecodeString("6ba7b8119dad11d180b400c04fd430c8")
	hash := sha1.New()
	_, _ = hash.Write(namespace)
	_, _ = hash.Write([]byte(fw.Name + "/" + fw.Version + "/" + fw.ContentDigest))
	b := hash.Sum(nil)[:16]
	b[6] = (b[6] & 0x0f) | 0x50
	b[8] = (b[8] & 0x3f) | 0x80
	id := fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
	catalog := Catalog{UUID: id, Metadata: Metadata{Title: fw.Name, LastModified: at.UTC(), Version: fw.Version, OSCALVersion: "1.1.2"}, Controls: []Control{}}
	for _, c := range fw.Controls {
		catalog.Controls = append(catalog.Controls, Control{ID: c.ID, Title: c.Name})
	}
	return Document{Catalog: catalog}, nil
}
func Import(data []byte) (*framework.Framework, error) {
	if len(data) > 4<<20 {
		return nil, fmt.Errorf("OSCAL catalog too large")
	}
	var document Document
	if err := json.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("decode OSCAL catalog: %w", err)
	}
	catalog := document.Catalog
	if catalog.UUID == "" || catalog.Metadata.Title == "" || catalog.Metadata.Version == "" || catalog.Metadata.OSCALVersion == "" {
		return nil, fmt.Errorf("OSCAL catalog identity is required")
	}
	fw := &framework.Framework{FrameworkSpec: framework.FrameworkSpec{Name: catalog.Metadata.Title, Version: catalog.Metadata.Version, Controls: []framework.Control{}}}
	seen := map[string]bool{}
	add := func(c Control) error {
		if c.ID == "" || c.Title == "" || seen[c.ID] || len(seen) >= 10000 {
			return fmt.Errorf("invalid or duplicate OSCAL control")
		}
		seen[c.ID] = true
		fw.Controls = append(fw.Controls, framework.Control{ID: c.ID, Name: c.Title, Severity: "medium", Mappings: []framework.Mapping{}})
		return nil
	}
	for _, c := range catalog.Controls {
		if err := add(c); err != nil {
			return nil, err
		}
	}
	var groups func([]Group, int) error
	groups = func(list []Group, depth int) error {
		if depth > 20 {
			return fmt.Errorf("OSCAL group nesting too deep")
		}
		for _, g := range list {
			for _, c := range g.Controls {
				if err := add(c); err != nil {
					return err
				}
			}
			if err := groups(g.Groups, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	if err := groups(catalog.Groups, 0); err != nil {
		return nil, err
	}
	if len(fw.Controls) == 0 {
		return nil, fmt.Errorf("OSCAL catalog has no controls")
	}
	return fw, nil
}
