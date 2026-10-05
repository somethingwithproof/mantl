package fleet

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTLSHandshakeRequiresTrustedEnrolledClient(t *testing.T) {
	pub, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	ca := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "fixture CA"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign}
	caDER, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, key)
	if err != nil {
		t.Fatal(err)
	}
	ca, err = x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca)
	makeClient := func(serial int64) tls.Certificate {
		public, private, genErr := ed25519.GenerateKey(rand.Reader)
		if genErr != nil {
			t.Fatal(genErr)
		}
		cert := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "fixture client"}, NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour), KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
		der, signErr := x509.CreateCertificate(rand.Reader, cert, ca, public, key)
		if signErr != nil {
			t.Fatal(signErr)
		}
		return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: private}
	}
	enrolled := makeClient(2)
	sum := sha256.Sum256(enrolled.Certificate[0])
	api := &Server{Enrollments: map[string]Scope{hex.EncodeToString(sum[:]): {Tenant: "tenant", Role: "viewer"}}, Index: &fixtureIndex{}}
	server := httptest.NewUnstartedServer(api)
	server.TLS = &tls.Config{MinVersion: tls.VersionTLS13, ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: roots}
	server.StartTLS()
	defer server.Close()
	invoke := func(cert *tls.Certificate) (*http.Response, error) {
		client := server.Client()
		transport := client.Transport.(*http.Transport).Clone()
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
		if cert != nil {
			transport.TLSClientConfig.Certificates = []tls.Certificate{*cert}
		}
		client = &http.Client{Transport: transport, Timeout: 5 * time.Second}
		defer transport.CloseIdleConnections()
		return client.Get(server.URL + "/v1/state")
	}
	if response, requestErr := invoke(nil); requestErr == nil {
		_ = response.Body.Close()
		t.Fatal("TLS accepted a client without a certificate")
	}
	other := makeClient(3)
	response, err := invoke(&other)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatal("trusted but unenrolled certificate accepted")
	}
	response, err = invoke(&enrolled)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("enrolled client rejected", response.StatusCode)
	}
}
