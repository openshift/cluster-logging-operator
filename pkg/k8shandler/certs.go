package k8shandler

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/utils"
	
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type certKeyPair struct {
	certName      string
	keyName       string
	componentName string
}

type certSecret struct {
	caNames           []string
	certs             []certKeyPair
	sessionSecretName *string
}

type certificate struct {
	cert     []byte
	key      []byte
	x509Cert *x509.Certificate
	privKey  *rsa.PrivateKey
}

type certCA struct {
	certificate
	serial big.Int
}

const rsaKeyLength = 4096
const caSecretName = `master-certs`
const caCertName = `ca.crt`
const caKeyName = `ca.key`
const caSerialName = `serial`
const caCN = `openshift-cluster-logging-signer`
const caNotAfterYears = 5

var ca *certCA
var sessionSecret = `session-secret`
var certs map[string]certificate
var secrets = map[string]*certSecret{
	`fluentd`: &certSecret{
		[]string{`ca-bundle.crt`},
		[]certKeyPair{
			certKeyPair{`tls.crt`, `tls.key`, `system.logging.fluentd`},
		},
		nil,
	},
	`elasticsearch`: &certSecret{
		[]string{`admin-ca`},
		[]certKeyPair{
			certKeyPair{`admin-cert`, `admin-key`, `system.admin`},
			certKeyPair{`logging-es.crt`, `logging-es.key`, `logging-es`},
			certKeyPair{`elasticsearch.crt`, `elasticsearch.key`, `elasticsearch`},
		},
		nil,
	},
	`curator`: &certSecret{
		[]string{`ca`, `ops-ca`},
		[]certKeyPair{
			certKeyPair{`cert`, `key`, `system.logging.curator`},
			certKeyPair{`ops-cert`, `ops-key`, `system.logging.curator`},
		},
		nil,
	},
	`kibana`: &certSecret{
		[]string{`ca`},
		[]certKeyPair{
			certKeyPair{`cert`, `key`, `system.logging.kibana`},
		},
		nil,
	},
	`kibana-proxy`: &certSecret{
		[]string{},
		[]certKeyPair{
			certKeyPair{`cert`, `key`, `kibana-internal`},
		},
		&sessionSecret,
	},
}

func pemEncodePrivateKey(privKey *rsa.PrivateKey) ([]byte, error) {
	pemBuffer := &bytes.Buffer{}
	if err := pem.Encode(pemBuffer, &pem.Block{
		Type:  `RSA PRIVATE KEY`,
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}); err != nil {
		return nil, err
	}
	return pemBuffer.Bytes(), nil
}

func pemDecodePrivateKey(keyBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != `RSA PRIVATE KEY` {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func pemEncodeCert(certDERBytes []byte) ([]byte, error) {
	pemBuffer := &bytes.Buffer{}
	if err := pem.Encode(pemBuffer, &pem.Block{Type: `CERTIFICATE`, Bytes: certDERBytes}); err != nil {
		return nil, err
	}
	return pemBuffer.Bytes(), nil
}

func pemDecodeCert(certPEMBytes []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certPEMBytes)
	if block == nil || block.Type != `CERTIFICATE` {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}
	return x509.ParseCertificate(block.Bytes)
}

func certHasExpired(cert *x509.Certificate) bool {
	return time.Now().After(cert.NotAfter)
}

func genCA() {

}

func validateCA(x509Cert *x509.Certificate, rsaPrivKey *rsa.PrivateKey) error {
	if certHasExpired(x509Cert) {
		return fmt.Errorf(`certificate has expired`)
	}
	rsaPubKey, ok := x509Cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return fmt.Errorf(`wrong public key type`)
	}
	if rsaPubKey.N.Cmp(rsaPrivKey.N) != 0 {
		return fmt.Errorf(`private key does not match public key`)
	}
	if !x509Cert.IsCA {
		return fmt.Errorf(`cert is not a CA`)
	}
	if x509Cert.Issuer.CommonName != caCN {
		return fmt.Errorf(`wrong issuer`)
	}
	if x509Cert.Subject.CommonName != caCN {
		return fmt.Errorf(`wrong subject`)
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) populateCA() error {
	if ca != nil {
		return nil
	}
	secret, err := clusterRequest.GetSecret(caSecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			secret = NewSecret(
				caSecretName,
				clusterRequest.Cluster.Namespace,
				map[string][]byte{},
			)
			utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))
			caPrivKey, err := rsa.GenerateKey(rand.Reader, rsaKeyLength)
			if err != nil {
				return err
			}
			caPubKeySHA1 := sha1.Sum(x509.MarshalPKCS1PublicKey(&caPrivKey.PublicKey))
			serial := big.NewInt(0)
			ca := &x509.Certificate{
				SerialNumber:          serial,
				SignatureAlgorithm:    x509.SHA512WithRSA,
				Subject:               pkix.Name{CommonName: caCN},
				NotBefore:             time.Now(),
				NotAfter:              time.Now().AddDate(caNotAfterYears, 0, 0),
				IsCA:                  true,
				KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
				BasicConstraintsValid: true,
				SubjectKeyId:          caPubKeySHA1[:],
				AuthorityKeyId:        caPubKeySHA1[:],
			}
			caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
			if err != nil {
				return err
			}
			caPEMBytes, err := pemEncodeCert(caBytes)
			if err != nil {
				return err
			}
			keyPEMBytes, err := pemEncodePrivateKey(caPrivKey)
			if err != nil {
				return err
			}
			serialBytes, err := serial.MarshalText()
			if err != nil {
				return err
			}
			secret.Data[caCertName] = caPEMBytes
			secret.Data[caKeyName] = keyPEMBytes
			secret.Data[caSerialName] = serialBytes
			if err = clusterRequest.Update(secret); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	var x509Cert *x509.Certificate
	certBytes, cert_ok := secret.Data[caCertName]
	if cert_ok {
		if x509Cert, err = pemDecodeCert(certBytes); err != nil {
			return err
		}
	}
	keyBytes, key_ok := secret.Data[caKeyName]
	var rsaKey *rsa.PrivateKey
	if key_ok {
		if rsaKey, err = pemDecodePrivateKey(keyBytes); err != nil {
			return err
		}
	}
	var serial big.Int
	serialBytes, serial_ok := secret.Data[caSerialName]
	if serial_ok {
		if err = serial.UnmarshalText(serialBytes); err != nil {
			return err
		}
	}
	if err = validateCA(x509Cert, rsaKey); err != nil {
		return err
	}
	ca = &certCA{
		certificate{
			certBytes,
			keyBytes,
			x509Cert,
			rsaKey,
		},
		serial,
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) checkCA(caCert []byte) ([]byte, error) {
	if err := clusterRequest.populateCA(); err != nil {
		return nil, err
	}
	if bytes.Compare(caCert, ca.cert) == 0 {
		return nil, nil
	}
	return ca.cert, nil
}

func (clusterRequest *ClusterLoggingRequest) checkCertAndKey(cert []byte, key []byte, componentName string) ([]byte, []byte, error) {
	compCert, ok := certs[componentName]
	if ok {
		if bytes.Compare(cert, compCert.cert) == 0 && bytes.Compare(key, compCert.key) == 0 {
			return nil, nil, nil
		}
	}
	return nil, nil, nil
}

func (clusterRequest *ClusterLoggingRequest) handleCAs(secret *core.Secret, keyNames []string) (dirty bool, err error) {
	for _, keyName := range keyNames {
		var newCACert []byte
		newCACert, err = clusterRequest.checkCA(secret.Data[keyName])
		if err != nil {
			return
		}
		if newCACert != nil {
			secret.Data[keyName] = newCACert
			dirty = true
		}
	}
	return
}

func (clusterRequest *ClusterLoggingRequest) handleCerts(secret *core.Secret, certs []certKeyPair) (dirty bool, err error) {
	for _, cert := range certs {
		newCert, newKey, err := clusterRequest.checkCertAndKey(
			secret.Data[cert.certName], secret.Data[cert.keyName], cert.componentName)
		if err != nil {
			return false, err
		}
		if newCert != nil {
			secret.Data[cert.certName] = newCert
			secret.Data[cert.keyName] = newKey
			dirty = true
		}
	}
	return
}

func handleSessionSecret(secret *core.Secret, sessionSecretName string) (dirty bool, err error) {
	// TODO syedriko: when does this need to be regen'd?
	const lenChars = 32
	const lenBytes = lenChars / 2
	sessionSecretBytes := make([]byte, lenBytes)
	_, err = rand.Read(sessionSecretBytes)
	if err != nil {
		return
	}
	sessionSecretChars := make([]byte, lenChars)
	hex.Encode(sessionSecretChars, sessionSecretBytes)
	secret.Data[sessionSecretName] = sessionSecretChars
	return true, nil
}

func (clusterRequest *ClusterLoggingRequest) ReconcileCertSecret(secretName string) (err error) {
	secret, err := clusterRequest.GetSecret(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			secret = NewSecret(
				secretName,
				clusterRequest.Cluster.Namespace,
				map[string][]byte{},
			)
			utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))
		} else {
			return err
		}
	}
	var dirtySecret bool
	secretDef, ok := secrets[secretName]
	if !ok {
		return fmt.Errorf("Unknown certificate secret definition: " + secretName)
	}
	dirtySecret, err = clusterRequest.handleCAs(secret, secretDef.caNames)
	if err != nil {
		return err
	}
	dirtySecret, err = clusterRequest.handleCerts(secret, secretDef.certs)
	if err != nil {
		return err
	}
	if secretDef.sessionSecretName != nil {
		dirtySecret, err = handleSessionSecret(secret, *secretDef.sessionSecretName)
		if err != nil {
			return err
		}
	}
	if dirtySecret {
		if err = clusterRequest.Update(secret); err != nil {
			return err
		}
	}
	return nil
}
