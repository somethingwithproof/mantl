package fleet

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Server struct {
	once        sync.Once
	gate        chan struct{}
	Enrollments map[string]Scope
	Index       Index
	Store       evidence.Store
	Now         func() time.Time
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.once.Do(func() { s.gate = make(chan struct{}, 8) })
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	default:
		http.Error(w, "request capacity exceeded", http.StatusTooManyRequests)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()
	r = r.WithContext(ctx)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if r.TLS == nil || len(r.TLS.VerifiedChains) == 0 || len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "authenticated client certificate required", http.StatusUnauthorized)
		return
	}
	sum := sha256.Sum256(r.TLS.PeerCertificates[0].Raw)
	scope, ok := s.Enrollments[hex.EncodeToString(sum[:])]
	if !ok || scope.Tenant == "" {
		http.Error(w, "enrollment unavailable", http.StatusForbidden)
		return
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/journals":
		if scope.Role != "collector" || scope.Cluster == "" {
			http.Error(w, "collector enrollment required", http.StatusForbidden)
			return
		}
		var ref evidence.ObjectRef
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&ref); err != nil {
			http.Error(w, "invalid journal reference", http.StatusBadRequest)
			return
		}
		now := time.Now().UTC()
		if s.Now != nil {
			now = s.Now().UTC()
		}
		events, err := Verify(r.Context(), s.Store, scope, ref, now)
		if err != nil {
			http.Error(w, "journal verification failed", http.StatusBadRequest)
			return
		}
		if err = s.Index.Put(r.Context(), scope, events, ref); err != nil {
			if err == ErrConflict {
				http.Error(w, "event identity conflict", http.StatusConflict)
			} else {
				http.Error(w, "index unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]int{"accepted": len(events)})
	case r.Method == http.MethodGet && (r.URL.Path == "/v1/events" || r.URL.Path == "/v1/state"):
		limit := 100
		if value := r.URL.Query().Get("limit"); value != "" {
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 || n > 500 {
				http.Error(w, "invalid page size", http.StatusBadRequest)
				return
			}
			limit = n
		}
		after := r.URL.Query().Get("after")
		if after != "" && !digestPattern.MatchString(after) {
			http.Error(w, "invalid cursor", http.StatusBadRequest)
			return
		}
		entries, err := s.Index.List(r.Context(), scope, limit, after, r.URL.Path == "/v1/state")
		if err != nil {
			http.Error(w, "index unavailable", http.StatusServiceUnavailable)
			return
		}
		if entries == nil {
			entries = []Entry{}
		}
		_ = json.NewEncoder(w).Encode(entries)
	default:
		http.NotFound(w, r)
	}
}
func ValidateEnrollments(enrollments map[string]Scope) error {
	if len(enrollments) == 0 {
		return fmt.Errorf("at least one enrollment is required")
	}
	clusters := map[string]string{}
	for fingerprint, scope := range enrollments {
		if !digestPattern.MatchString(fingerprint) || !scopePattern.MatchString(scope.Tenant) || scope.Cluster != "" && !scopePattern.MatchString(scope.Cluster) || (scope.Role != "collector" && scope.Role != "viewer") || scope.Role == "collector" && scope.Cluster == "" {
			return fmt.Errorf("invalid fleet enrollment")
		}
		if scope.Cluster != "" {
			if owner, ok := clusters[scope.Cluster]; ok && owner != scope.Tenant {
				return fmt.Errorf("cluster enrolled in multiple tenants")
			}
			clusters[scope.Cluster] = scope.Tenant
		}
	}
	return nil
}
