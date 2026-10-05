package fleet

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/thomasvincent/mantl/pkg/evidence"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Client struct {
	URL  string
	HTTP *http.Client
}

func NewClient(endpoint, cert, key, ca string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" && u.Path != "/" {
		return nil, fmt.Errorf("fleet endpoint must be an HTTPS origin without credentials")
	}
	pair, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("load fleet client identity: %w", err)
	}
	data, err := os.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("invalid fleet server CA")
	}
	u.Path = ""
	return &Client{URL: u.String(), HTTP: &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS13, Certificates: []tls.Certificate{pair}, RootCAs: roots}, MaxIdleConns: 4, IdleConnTimeout: 30 * time.Second}}}, nil
}
func (c *Client) Publish(ctx context.Context, ref evidence.ObjectRef) error {
	data, err := json.Marshal(ref)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+"/v1/journals", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("publish fleet journal: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	_, err = io.Copy(io.Discard, io.LimitReader(response.Body, 8192))
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("fleet journal rejected: HTTP %d", response.StatusCode)
	}
	return nil
}
func (c *Client) State(ctx context.Context) ([]Entry, error) {
	var all []Entry
	after := ""
	for page := 0; page < 200; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/v1/state?limit=500&after="+after, nil)
		if err != nil {
			return nil, err
		}
		response, err := c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
		if response.StatusCode != http.StatusOK {
			_ = response.Body.Close()
			return nil, fmt.Errorf("fleet state rejected: HTTP %d", response.StatusCode)
		}
		var entries []Entry
		err = json.NewDecoder(io.LimitReader(response.Body, 8<<20)).Decode(&entries)
		_ = response.Body.Close()
		if err != nil {
			return nil, err
		}
		all = append(all, entries...)
		if len(entries) < 500 {
			return all, nil
		}
		next := entries[len(entries)-1].Event.ID
		if !digestPattern.MatchString(next) || next <= after {
			return nil, fmt.Errorf("invalid fleet pagination cursor")
		}
		after = next
	}
	return nil, fmt.Errorf("fleet state exceeds 100000 entries; narrow the enrolled cluster scope")
}
func (c *Client) Journal(ctx context.Context, store evidence.Store, cluster string, events []Event, now time.Time) error {
	data, err := Canonical(cluster, events)
	if err != nil {
		return err
	}
	ref, err := store.Put(ctx, Prefix(cluster), data, now)
	if err != nil {
		return err
	}
	return c.Publish(ctx, ref)
}
