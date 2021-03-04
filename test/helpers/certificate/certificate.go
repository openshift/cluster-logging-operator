package certificate

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/openshift/cluster-logging-operator/test"
)

// CertKey holds a certificate struct, private key and DER encoded bytes.
type CertKey struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
	DERBytes    []byte
}

// New CertKey based on template and signed by signer, or self-signed if signer is nil.
func New(template *x509.Certificate, signer *CertKey) *CertKey {
	ck := &CertKey{Certificate: template}
	var err error
	ck.PrivateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	test.Must(err)
	if signer == nil {
		signer = ck // Self signed
	}
	ck.DERBytes, err = x509.CreateCertificate(rand.Reader, ck.Certificate, signer.Certificate, &ck.PrivateKey.PublicKey, signer.PrivateKey)
	test.Must(err)
	return ck
}

func (ck *CertKey) CertificatePEM() []byte {
	var b bytes.Buffer
	test.Must(pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: ck.DERBytes}))
	return b.Bytes()
}

func (ck *CertKey) PrivateKeyPEM() []byte {
	var b bytes.Buffer
	test.Must(pem.Encode(&b, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(ck.PrivateKey)}))
	return b.Bytes()
}

// ServerTLSConf returns TLS configuration for a server using this certificate.
// If clientCA is not nil, we also enable client authentication using clientCA.
func (ck *CertKey) ServerTLSConf(clientCA *CertKey) *tls.Config {
	cert, err := tls.X509KeyPair(ck.CertificatePEM(), ck.PrivateKeyPEM())
	test.Must(err)
	conf := &tls.Config{Certificates: []tls.Certificate{cert}}
	if clientCA != nil {
		conf.ClientCAs = x509.NewCertPool()
		conf.ClientCAs.AppendCertsFromPEM(clientCA.CertificatePEM())
		conf.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return conf
}

// ClientTLSConf returns TLS configuration for a client using this cert as a CA.
// If clientCert is not nil, client will authenticate with cert.
func (ck *CertKey) ClientTLSConf(clientCert *CertKey) *tls.Config {
	certpool := x509.NewCertPool()
	certpool.AppendCertsFromPEM(ck.CertificatePEM())
	conf := &tls.Config{RootCAs: certpool}
	if clientCert != nil {
		cert, err := tls.X509KeyPair(clientCert.CertificatePEM(), clientCert.PrivateKeyPEM())
		test.Must(err)
		conf.Certificates = []tls.Certificate{cert}
	}
	return conf
}

// Client returns a HTTP client using this cert as CA.
// If clientCert is not nil, client will authenticate with cert.
func (ck *CertKey) Client(clientCert *CertKey) *http.Client {
	return &http.Client{Transport: &http.Transport{TLSClientConfig: ck.ClientTLSConf(clientCert)}}
}

// StartServer returns a started httptest.Server using this cert.
// If clientCA is not nil, enable client authentication.
func (ck *CertKey) StartServer(handler http.Handler, clientCA *CertKey) *httptest.Server {
	server := httptest.NewUnstartedServer(handler)
	server.TLS = ck.ServerTLSConf(clientCA)
	server.StartTLS()
	return server
}

// NewCA creates a new dummy CA cert signed by signer, or self-signed if signer is nil.
func NewCA(signer *CertKey, org string) *CertKey {
	return New(&x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			Organization: []string{org},
			// Country:       []string{"US"},
			// Province:      []string{""},
			// Locality:      []string{"San Francisco"},
			// StreetAddress: []string{"Golden Gate Bridge"},
			// PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}, signer)
}

// NewCert creates a dummy server cert signed by signer, or self-signed if signer is nil.
// The addrs list can contain strings (DNS names) or net.IP addresses, if addrs
// is empty will use "localhost", v4 and v6 loopback
func NewCert(signer *CertKey, org string, addrs ...interface{}) *CertKey {
	var (
		dns []string
		ips []net.IP
	)
	if len(addrs) == 0 {
		addrs = []interface{}{"localhost", net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	}
	for _, a := range addrs {
		switch a := a.(type) {
		case string:
			dns = append(dns, a)
		case net.IP:
			ips = append(ips, a)
		default:
			test.Must(fmt.Errorf("expected string or net.IP, got (%T)%#v", a, a))
		}
	}
	return New(&x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject:      pkix.Name{Organization: []string{org}},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,

		IPAddresses: ips,
		DNSNames:    dns,
	}, signer)
}

// NewServer creates a dummy client cert signed by signer, or self-signed if signer is nil.
func NewClient(signer *CertKey, org string) *CertKey {
	// Its the same as a server cert but with no DNS/IP addrs.
	return NewCert(signer, org)
}
