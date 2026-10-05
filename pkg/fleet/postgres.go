package fleet

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/thomasvincent/mantl/pkg/evidence"
)

type Postgres struct{ DB *sql.DB }

func (p *Postgres) Migrate(ctx context.Context) error {
	var privileged bool
	if err := p.DB.QueryRowContext(ctx, "SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname=current_user").Scan(&privileged); err != nil {
		return err
	}
	if privileged {
		return fmt.Errorf("fleet database role must not bypass row-level security")
	}
	_, err := p.DB.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS mantl_events (
 tenant_id text NOT NULL,cluster_id text NOT NULL,id text NOT NULL,
 kind text NOT NULL,resource_uid text NOT NULL,observed_at timestamptz NOT NULL,
 payload jsonb NOT NULL,payload_hash text NOT NULL,source jsonb NOT NULL,
 PRIMARY KEY(tenant_id,cluster_id,id));
 CREATE INDEX IF NOT EXISTS mantl_event_state ON mantl_events(tenant_id,cluster_id,kind,resource_uid,observed_at DESC);
 ALTER TABLE mantl_events ENABLE ROW LEVEL SECURITY;
 ALTER TABLE mantl_events FORCE ROW LEVEL SECURITY;
 DROP POLICY IF EXISTS mantl_tenant_scope ON mantl_events;
 CREATE POLICY mantl_tenant_scope ON mantl_events USING (
 tenant_id=current_setting('mantl.tenant_id',true) AND
 (current_setting('mantl.cluster_id',true)='' OR cluster_id=current_setting('mantl.cluster_id',true)))
 WITH CHECK (tenant_id=current_setting('mantl.tenant_id',true) AND cluster_id=current_setting('mantl.cluster_id',true));`)
	return err
}
func (p *Postgres) transaction(ctx context.Context, scope Scope) (*sql.Tx, error) {
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, "SELECT set_config('mantl.tenant_id',$1,true),set_config('mantl.cluster_id',$2,true)", scope.Tenant, scope.Cluster); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	return tx, nil
}
func (p *Postgres) Put(ctx context.Context, scope Scope, events []Event, source evidence.ObjectRef) error {
	tx, err := p.transaction(ctx, scope)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	ref, err := json.Marshal(source)
	if err != nil {
		return err
	}
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}
		hash := evidence.Hash(b)
		_, err = tx.ExecContext(ctx, `INSERT INTO mantl_events(tenant_id,cluster_id,id,kind,resource_uid,observed_at,payload,payload_hash,source)
 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) ON CONFLICT(tenant_id,cluster_id,id) DO NOTHING`, scope.Tenant, scope.Cluster, e.ID, e.Kind, e.ResourceUID, e.ObservedAt, b, hash, ref)
		if err != nil {
			return err
		}
		var existing string
		if err = tx.QueryRowContext(ctx, "SELECT payload_hash FROM mantl_events WHERE tenant_id=$1 AND cluster_id=$2 AND id=$3", scope.Tenant, scope.Cluster, e.ID).Scan(&existing); err != nil {
			return err
		}
		if existing != hash {
			return ErrConflict
		}
	}
	return tx.Commit()
}
func (p *Postgres) List(ctx context.Context, scope Scope, limit int, after string, state bool) ([]Entry, error) {
	tx, err := p.transaction(ctx, scope)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	query := `SELECT cluster_id,payload,source FROM mantl_events WHERE tenant_id=$1 AND ($2='' OR cluster_id=$2) AND id>$3 ORDER BY id LIMIT $4`
	if state {
		query = `SELECT cluster_id,payload,source FROM (
 SELECT DISTINCT ON(cluster_id,kind,resource_uid) cluster_id,payload,source,id,observed_at
 FROM mantl_events WHERE tenant_id=$1 AND ($2='' OR cluster_id=$2)
 ORDER BY cluster_id,kind,resource_uid,observed_at DESC,id DESC) current_state WHERE id>$3 ORDER BY id LIMIT $4`
	}
	rows, err := tx.QueryContext(ctx, query, scope.Tenant, scope.Cluster, after, limit)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for rows.Next() {
		var cluster string
		var payload, ref []byte
		if err = rows.Scan(&cluster, &payload, &ref); err != nil {
			_ = rows.Close()
			return nil, err
		}
		entry := Entry{Scope: Scope{Tenant: scope.Tenant, Cluster: cluster, Role: scope.Role}}
		if err = json.Unmarshal(payload, &entry.Event); err != nil {
			_ = rows.Close()
			return nil, err
		}
		if err = json.Unmarshal(ref, &entry.Source); err != nil {
			_ = rows.Close()
			return nil, err
		}
		entries = append(entries, entry)
	}
	readErr := rows.Err()
	closeErr := rows.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return entries, tx.Commit()
}
