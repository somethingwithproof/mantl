package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"github.com/thomasvincent/mantl/pkg/fleet"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run() error {
	var addr, cert, key, ca, enrollment, bucket string
	flag.StringVar(&addr, "listen", ":8443", "TLS listener")
	flag.StringVar(&cert, "tls-cert", "", "Server certificate path")
	flag.StringVar(&key, "tls-key", "", "Server key path")
	flag.StringVar(&ca, "client-ca", "", "Trusted client CA path")
	flag.StringVar(&enrollment, "enrollments", "", "Reviewed certificate fingerprint-to-scope JSON file")
	flag.StringVar(&bucket, "evidence-bucket", "", "Immutable metadata journal bucket")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	data, err := os.ReadFile(enrollment)
	if err != nil {
		return err
	}
	var scopes map[string]fleet.Scope
	if err = json.Unmarshal(data, &scopes); err != nil {
		return err
	}
	if err = fleet.ValidateEnrollments(scopes); err != nil {
		return err
	}
	caBytes, err := os.ReadFile(ca)
	if err != nil {
		return err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return fmt.Errorf("invalid client CA")
	}
	dsn := os.Getenv("MANTL_FLEET_DATABASE_URL")
	if dsn == "" || bucket == "" {
		return fmt.Errorf("database runtime credentials and evidence bucket are required")
	}
	databaseConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("configure fleet database")
	}
	if databaseConfig.TLSConfig == nil || databaseConfig.TLSConfig.InsecureSkipVerify || len(databaseConfig.Fallbacks) > 0 {
		return fmt.Errorf("fleet database requires verified TLS; use sslmode=verify-full")
	}
	db := stdlib.OpenDB(*databaseConfig)
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	startup, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	index := &fleet.Postgres{DB: db}
	if err = index.Migrate(startup); err != nil {
		return fmt.Errorf("prepare fleet index: %w", err)
	}
	aws, err := config.LoadDefaultConfig(startup)
	if err != nil {
		return fmt.Errorf("configure immutable journal verification")
	}
	handler := &fleet.Server{Enrollments: scopes, Index: index, Store: &evidence.S3Store{Client: s3.NewFromConfig(aws), Bucket: bucket}}
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384, TLSConfig: &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: pool}}
	done := make(chan error, 1)
	go func() { done <- server.ListenAndServeTLS(cert, key) }()
	select {
	case err = <-done:
		if err != http.ErrServerClosed {
			return err
		}
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		slog.Error("fleet server stopped", "error", err)
		os.Exit(1)
	}
}
