// SPDX-License-Identifier: Apache-2.0

//go:build integration

package fleet

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"os"
	"testing"
	"time"
)

func TestPostgresTenantIsolationAndReplay(t *testing.T) {
	dsn := os.Getenv("MANTL_TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("integration database must be explicitly configured")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal("database configuration failed")
	}
	defer func() { _ = db.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	index := &Postgres{DB: db}
	if err = index.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	stamp := fmt.Sprint(time.Now().UnixNano())
	a := Scope{Tenant: "a-" + stamp, Cluster: "cluster-a", Role: "collector"}
	b := Scope{Tenant: "b-" + stamp, Cluster: "cluster-b", Role: "collector"}
	event := Event{ID: evidence.Hash([]byte("same-event-id")), Kind: "Finding", ResourceUID: "uid", Namespace: "team", Name: "finding", ObservedAt: time.Now().UTC().Truncate(time.Second), Result: "fail"}
	ref := evidence.ObjectRef{URI: "s3://fixture/journal", Hash: evidence.Hash([]byte("journal")), Version: "1"}
	if err = index.Put(ctx, a, []Event{event}, ref); err != nil {
		t.Fatal(err)
	}
	if err = index.Put(ctx, a, []Event{event}, ref); err != nil {
		t.Fatal("idempotent replay failed", err)
	}
	if err = index.Put(ctx, b, []Event{event}, ref); err != nil {
		t.Fatal(err)
	}
	entries, err := index.List(ctx, a, 100, "", false)
	if err != nil || len(entries) != 1 || entries[0].Scope.Tenant != a.Tenant {
		t.Fatal("cross-tenant data exposed", err)
	}
	// Forced RLS must also protect an accidental unscoped SQL query.
	tx, err := index.transaction(ctx, a)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err = tx.QueryRowContext(ctx, "SELECT count(*) FROM mantl_events").Scan(&count); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err = tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("RLS did not restrict unscoped query")
	}
	event.Result = "resolved"
	if err = index.Put(ctx, a, []Event{event}, ref); err != ErrConflict {
		t.Fatal("changed replay accepted", err)
	}
	event.ID = evidence.Hash([]byte("new-event"))
	event.ObservedAt = event.ObservedAt.Add(time.Second)
	if err = index.Put(ctx, a, []Event{event}, ref); err != nil {
		t.Fatal(err)
	}
	state, err := index.List(ctx, a, 100, "", true)
	if err != nil || len(state) != 1 || state[0].Event.Result != "resolved" {
		t.Fatal("latest state incorrect", err)
	}
}
