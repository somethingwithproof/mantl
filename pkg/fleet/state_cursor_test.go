// SPDX-License-Identifier: Apache-2.0

package fleet

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStateQueryPagesByResourceIdentity(t *testing.T) {
	if strings.Contains(stateListQuery, "id>$3") || !strings.Contains(stateListQuery, "(cluster_id,kind,resource_uid)") {
		t.Fatal("state page still follows mutable event ids")
	}
	if !stateCursorBefore("a/Finding/a", "a./Finding/z") || "a/Finding/a" < "a./Finding/z" {
		t.Fatal("cursor order disagrees with cluster, kind, resource order")
	}
	same := StateCursor("cluster-a", "Finding", "uid")
	if !ValidStateCursor(same) || stateCursorBefore(same, same) {
		t.Fatal("resource cursor changed or compared a resource with itself")
	}
}

func TestStatePagesPastReorderedEventIDs(t *testing.T) {
	var cursors []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cursors = append(cursors, r.URL.Query().Get("after"))
		entries := page("cluster-a", "Finding", "m", strings.Repeat("f", 64))
		if len(cursors) > 1 {
			entries = []Entry{{Scope: Scope{Cluster: "cluster-a"}, Event: Event{ID: strings.Repeat("0", 64), Kind: "Finding", ResourceUID: "z"}}}
		}
		_ = json.NewEncoder(w).Encode(entries)
	}))
	defer server.Close()
	entries, err := (&Client{URL: server.URL, HTTP: server.Client()}).State(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 501 || len(cursors) != 2 || cursors[1] != "cluster-a/Finding/m" {
		t.Fatalf("pages: %d cursors %v", len(entries), cursors)
	}
}

func page(cluster, kind, uid, id string) []Entry {
	entries := make([]Entry, 500)
	for i := range entries {
		resource := uid
		if i+1 < len(entries) {
			resource = uid + "-early"
		}
		entries[i] = Entry{Scope: Scope{Cluster: cluster}, Event: Event{ID: id, Kind: kind, ResourceUID: resource}}
	}
	return entries
}
