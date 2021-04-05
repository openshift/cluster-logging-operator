package certificate

import (
	"bytes"
	"crypto/x509"
	"strconv"
	"time"

	"github.com/openshift/cluster-logging-operator/test"
	"github.com/pavel-v-chernykh/keystore-go/v4"
)

func JKSKeyStore(certKey *CertKey, password string) []byte {
	ks := keystore.New()
	keyPKCS8Bytes, err := x509.MarshalPKCS8PrivateKey(certKey.PrivateKey)
	test.Must(err)
	pke := keystore.PrivateKeyEntry{
		CreationTime: time.Now(),
		PrivateKey:   keyPKCS8Bytes,
		CertificateChain: []keystore.Certificate{
			{
				Type:    "X509",
				Content: certKey.DERBytes,
			},
		},
	}
	err = ks.SetPrivateKeyEntry("alias", pke, []byte(password))
	test.Must(err)
	var buf bytes.Buffer
	err = ks.Store(&buf, []byte(password))
	test.Must(err)
	return buf.Bytes()
}

func JKSTrustStore(certKeys []*CertKey, password string) []byte {
	ks := keystore.New()
	for i := range certKeys {
		err := ks.SetTrustedCertificateEntry(
			strconv.Itoa(i),
			keystore.TrustedCertificateEntry{
				CreationTime: time.Now(),
				Certificate: keystore.Certificate{
					Type:    "X509",
					Content: certKeys[i].DERBytes,
				},
			},
		)
		test.Must(err)
	}
	var buf bytes.Buffer
	err := ks.Store(&buf, []byte(password))
	test.Must(err)
	return buf.Bytes()
}
