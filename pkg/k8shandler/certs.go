package k8shandler

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/openshift/cluster-logging-operator/pkg/utils"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

type compCertDef struct {
	certName      string
	keyName       string
	componentName string
}

type certSecretDef struct {
	caNames           []string
	compCertDefs      []compCertDef
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
	serial     *big.Int
	pubKeySHA1 []byte
}

type x509v3Ext struct {
	dns []string
	ips []net.IP
	oid bool
}

const rsaKeyLength = 4096
const caSecretName = `master-certs`
const caCertName = `ca.crt`
const caKeyName = `ca.key`
const caSerialName = `serial`
const caCN = `openshift-cluster-logging-signer`
const caNotAfterYears = 5
const compNotAfterYears = 2
const nsPlaceholder = `${NAMESPACE}`

var bigOne = big.NewInt(1)
var cloCA *certCA
var cloCAMutex sync.Mutex
var sessionSecret = `session-secret`
var certs map[string]certificate
var certsMutex sync.Mutex
var certSecretDefs = map[string]*certSecretDef{
	`fluentd`: &certSecretDef{
		[]string{`ca-bundle.crt`},
		[]compCertDef{
			compCertDef{`tls.crt`, `tls.key`, `system.logging.fluentd`},
		},
		nil,
	},
	`elasticsearch`: &certSecretDef{
		[]string{`admin-ca`},
		[]compCertDef{
			compCertDef{`admin-cert`, `admin-key`, `system.admin`},
			compCertDef{`logging-es.crt`, `logging-es.key`, `logging-es`},
			compCertDef{`elasticsearch.crt`, `elasticsearch.key`, `elasticsearch`},
		},
		nil,
	},
	`curator`: &certSecretDef{
		[]string{`ca`, `ops-ca`},
		[]compCertDef{
			compCertDef{`cert`, `key`, `system.logging.curator`},
			compCertDef{`ops-cert`, `ops-key`, `system.logging.curator`},
		},
		nil,
	},
	`kibana`: &certSecretDef{
		[]string{`ca`},
		[]compCertDef{
			compCertDef{`cert`, `key`, `system.logging.kibana`},
		},
		nil,
	},
	`kibana-proxy`: &certSecretDef{
		[]string{},
		[]compCertDef{
			compCertDef{`cert`, `key`, `kibana-internal`},
		},
		&sessionSecret,
	},
}

var extensions = map[string]x509v3Ext{
	`kibana-internal`: x509v3Ext{
		[]string{`kibana`},
		[]net.IP{},
		false,
	},
	`elasticsearch`: x509v3Ext{
		[]string{`localhost`,
			`elasticsearch`,
			`elasticsearch.cluster.local`,
			`elasticsearch.` + nsPlaceholder + `.svc`,
			`elasticsearch.` + nsPlaceholder + `.svc.cluster.local`,
			`elasticsearch-cluster`,
			`elasticsearch-cluster.cluster.local`,
			`elasticsearch-cluster.` + nsPlaceholder + `.svc`,
			`elasticsearch-cluster.` + nsPlaceholder + `.svc.cluster.local`,
		},
		[]net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		true,
	},
	`logging-es`: x509v3Ext{
		[]string{`localhost`,
			`elasticsearch`,
			`elasticsearch.cluster.local`,
			`elasticsearch.` + nsPlaceholder + `.svc`,
			`elasticsearch.` + nsPlaceholder + `.svc.cluster.local`,
		},
		[]net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		false,
	},
}

var extInit sync.Once
var oidExt pkix.Extension = pkix.Extension{
	Id: asn1.ObjectIdentifier{2, 5, 29, 17},
	Value: func() []byte {
		oidExtASN1Bytes, _ := asn1.Marshal([]asn1.RawValue{
			{FullBytes: []byte{0x88, 0x05, 0x2A, 0x03, 0x04, 0x05, 0x05}},
		})
		return oidExtASN1Bytes
	}(),
}

func initExtensions(ns string) {
	for _, v := range extensions {
		for i := range v.dns {
			strings.ReplaceAll(v.dns[i], nsPlaceholder, ns)
		}
	}
}

func cloneByteSlice(s []byte) []byte {
	r := make([]byte, len(s))
	copy(r, s)
	return r
}

func (lhs *certificate) equals(rhs *certificate) bool {
	return bytes.Equal(lhs.cert, rhs.cert) && bytes.Equal(lhs.key, rhs.key)
}

func (lhs *certCA) equals(rhs *certCA) bool {
	return lhs.certificate.equals(&rhs.certificate) && lhs.serial.Cmp(rhs.serial) == 0
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

func genCA() (*certCA, error) {
	caPrivKey, err := rsa.GenerateKey(rand.Reader, rsaKeyLength)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	caPEMBytes, err := pemEncodeCert(caBytes)
	if err != nil {
		return nil, err
	}
	keyPEMBytes, err := pemEncodePrivateKey(caPrivKey)
	if err != nil {
		return nil, err
	}
	return &certCA{
		certificate{
			caPEMBytes,
			keyPEMBytes,
			ca,
			caPrivKey,
		},
		serial,
		caPubKeySHA1[:],
	}, nil
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

func validateCASecret(secret *core.Secret) (*certCA, error) {
	var x509Cert *x509.Certificate
	var err error
	certBytes, cert_ok := secret.Data[caCertName]
	if cert_ok {
		if x509Cert, err = pemDecodeCert(certBytes); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf(`Missing ` + caCertName + ` key from secret ` + caSecretName)
	}
	var rsaKey *rsa.PrivateKey
	keyBytes, key_ok := secret.Data[caKeyName]
	if key_ok {
		if rsaKey, err = pemDecodePrivateKey(keyBytes); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf(`Missing ` + caKeyName + ` key from secret ` + caSecretName)
	}
	pubKeySHA1 := sha1.Sum(x509.MarshalPKCS1PublicKey(&rsaKey.PublicKey))
	var serial *big.Int
	serialBytes, serial_ok := secret.Data[caSerialName]
	if serial_ok {
		if err = serial.UnmarshalText(serialBytes); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf(`Missing ` + caSerialName + ` key from secret ` + caSecretName)
	}
	if err = validateCA(x509Cert, rsaKey); err != nil {
		return nil, err
	}
	return &certCA{
		certificate{
			certBytes,
			keyBytes,
			x509Cert,
			rsaKey,
		},
		serial,
		pubKeySHA1[:],
	}, nil
}

func (clusterRequest *ClusterLoggingRequest) newCertSecret(name string) *core.Secret {
	secret := NewSecret(
		name,
		clusterRequest.Cluster.Namespace,
		map[string][]byte{},
	)
	utils.AddOwnerRefToObject(secret, utils.AsOwner(clusterRequest.Cluster))
	return secret
}

func (clusterRequest *ClusterLoggingRequest) populateCASecret(secret *core.Secret, cert []byte, key []byte, serial []byte) (err error) {
	secret.Data[caCertName] = cert
	secret.Data[caKeyName] = key
	secret.Data[caSerialName] = serial
	if err = clusterRequest.Update(secret); err != nil {
		return err
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) populateCA() error {
	secret, err := clusterRequest.GetSecret(caSecretName)
	if err != nil {
		if errors.IsNotFound(err) {
			// the CA secret is not there
			secret = clusterRequest.newCertSecret(caSecretName)
			var cert, key, serial []byte
			if cert, key, serial, err = func() ([]byte, []byte, []byte, error) {
				cloCAMutex.Lock()
				defer cloCAMutex.Unlock()
				if cloCA == nil {
					// no CLO CA, generate it
					if cloCA, err = genCA(); err != nil {
						return nil, nil, nil, err
					}
				}
				return cloneByteSlice(cloCA.cert), cloneByteSlice(cloCA.key), []byte(cloCA.serial.Text(10)), nil
			}(); err != nil {
				return err
			}
			// populate the CA secret with the CLO CA
			if err = clusterRequest.populateCASecret(secret, cert, key, serial); err != nil {
				return err
			}
		} else {
			// unable to get CA secret for reason other than NotFound
			return err
		}
	} else {
		// found CA secret
		var cert, key, serial []byte
		if cert, key, serial, err = func() ([]byte, []byte, []byte, error) {
			cloCAMutex.Lock()
			defer cloCAMutex.Unlock()
			if cloCA != nil {
				// CLO has CA
				tempCA, err := validateCASecret(secret)
				if err != nil {
					// invalid CA secret, populate CA secret from CLO CA
					secret = clusterRequest.newCertSecret(caSecretName)
					return cloneByteSlice(cloCA.cert), cloneByteSlice(cloCA.key), []byte(cloCA.serial.Text(10)), nil
				} else {
					// valid CA secret, check if it's the same as CLO CA
					if !cloCA.equals(tempCA) {
						// it's not, populate the secret from CLO CA
						secret = clusterRequest.newCertSecret(caSecretName)
						return cloneByteSlice(cloCA.cert), cloneByteSlice(cloCA.key), []byte(cloCA.serial.Text(10)), nil
					} else {
						// CLO CA is the same as what's in the CA secret, do nothing
					}
				}
			} else {
				// no CLO CA, populate CLO CA from the CA secret if it is valid
				cloCA, err = validateCASecret(secret)
				if err != nil {
					// generate new CLO CA and populate the CA secret with it
					if cloCA, err = genCA(); err != nil {
						return nil, nil, nil, err
					}
					return cloneByteSlice(cloCA.cert), cloneByteSlice(cloCA.key), []byte(cloCA.serial.Text(10)), nil
				}
			}
			return nil, nil, nil, nil
		}(); err != nil {
			return err
		}
		if cert != nil {
			// populate the CA secret with the CLO CA
			if err = clusterRequest.populateCASecret(secret, cert, key, serial); err != nil {
				return err
			}
		}
	}
	return nil
}

func genCert(componentName string) (ret certificate, err error) {
	if ret.privKey, err = rsa.GenerateKey(rand.Reader, rsaKeyLength); err != nil {
		return
	}
	pubKeySHA1 := sha1.Sum(x509.MarshalPKCS1PublicKey(&ret.privKey.PublicKey))
	cloCAMutex.Lock()
	defer cloCAMutex.Unlock()
	ret.x509Cert = &x509.Certificate{
		SerialNumber:       cloCA.serial.Add(cloCA.serial, bigOne),
		SignatureAlgorithm: x509.SHA512WithRSA,
		Subject: pkix.Name{
			Organization:       []string{`Logging`},
			OrganizationalUnit: []string{`OpenShift`},
			CommonName:         componentName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(compNotAfterYears, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		BasicConstraintsValid: true,
		SubjectKeyId:          pubKeySHA1[:],
		AuthorityKeyId:        cloCA.pubKeySHA1[:],
	}

	ext := extensions[componentName]
	if ext.dns != nil {
		ret.x509Cert.DNSNames = ext.dns
	}
	if ext.ips != nil {
		ret.x509Cert.IPAddresses = ext.ips
	}
	var certBytes []byte
	certBytes, err = x509.CreateCertificate(rand.Reader, ret.x509Cert, cloCA.x509Cert, &ret.privKey.PublicKey, cloCA.privKey)
	if err != nil {
		return
	}
	if ext.oid {
		ret.x509Cert, err = x509.ParseCertificate(certBytes)
		if err != nil {
			return
		}
		ret.x509Cert.ExtraExtensions = ret.x509Cert.Extensions
		ret.x509Cert.ExtraExtensions = append(
			ret.x509Cert.ExtraExtensions,
			oidExt,
		)
		certBytes, err = x509.CreateCertificate(rand.Reader, ret.x509Cert, cloCA.x509Cert, &ret.privKey.PublicKey, cloCA.privKey)
		if err != nil {
			return
		}
	}
	ret.cert, err = pemEncodeCert(certBytes)
	if err != nil {
		return
	}
	ret.key, err = pemEncodePrivateKey(ret.privKey)
	if err != nil {
		return
	}
	return
}

func (clusterRequest *ClusterLoggingRequest) checkCertAndKey(cert []byte, key []byte, componentName string) ([]byte, []byte, error) {
	certsMutex.Lock()
	defer certsMutex.Unlock()
	var compCert certificate
	compCert, found := certs[componentName]
	if found {
		if bytes.Equal(cert, compCert.cert) && bytes.Equal(key, compCert.key) {
			return nil, nil, nil
		} else {
			return cloneByteSlice(compCert.cert), cloneByteSlice(compCert.key), nil
		}
	} else {
		if err := func() (err error) {
			certsMutex.Unlock()
			defer certsMutex.Lock()
			compCert, err = genCert(componentName)
			return err
		}(); err != nil {
			return nil, nil, err
		}
		compCertInMap, found := certs[componentName]
		if found {
			// another goroutine beat us to it, should be unlikely
			return cloneByteSlice(compCertInMap.cert), cloneByteSlice(compCertInMap.key), nil
		} else {
			certs[componentName] = compCert
			return cloneByteSlice(compCert.cert), cloneByteSlice(compCert.key), nil
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) reconcileCAs(secret *core.Secret, keyNames []string) (dirty bool, err error) {
	cloCAMutex.Lock()
	defer cloCAMutex.Unlock()
	for _, keyName := range keyNames {
		if !bytes.Equal(secret.Data[keyName], cloCA.cert) {
			secret.Data[keyName] = cloneByteSlice(cloCA.cert)
			dirty = true
		}
	}
	return
}

func (clusterRequest *ClusterLoggingRequest) reconcileCerts(secret *core.Secret, compCertDefs []compCertDef) (dirty bool, err error) {
	for _, certDef := range compCertDefs {
		newCert, newKey, err := clusterRequest.checkCertAndKey(
			secret.Data[certDef.certName], secret.Data[certDef.keyName], certDef.componentName)
		if err != nil {
			return false, err
		}
		if newCert != nil {
			secret.Data[certDef.certName] = newCert
			secret.Data[certDef.keyName] = newKey
			dirty = true
		}
	}
	return
}

func reconcileSessionSecret(secret *core.Secret, sessionSecretName string) (dirty bool, err error) {
	// TODO syedriko: when does this need to be regen'd?
	const lenChars = 32
	const lenBytes = lenChars / 2
	var sessionSecretBytes [lenBytes]byte
	sessionSecretBytesSlice := sessionSecretBytes[:]
	if _, err = rand.Read(sessionSecretBytesSlice); err != nil {
		return
	}
	var sessionSecretChars [lenChars]byte
	sessionSecretCharsSlice := sessionSecretChars[:]
	hex.Encode(sessionSecretCharsSlice, sessionSecretBytesSlice)
	secret.Data[sessionSecretName] = sessionSecretCharsSlice
	return true, nil
}

func (clusterRequest *ClusterLoggingRequest) ReconcileCertSecret(secretName string) error {
	extInit.Do(func() { initExtensions(clusterRequest.Cluster.Namespace) })
	if err := clusterRequest.populateCA(); err != nil {
		return err
	}
	secret, err := clusterRequest.GetSecret(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			secret = clusterRequest.newCertSecret(secretName)
		} else {
			return err
		}
	}
	var dirtySecret bool
	secretDef, defined := certSecretDefs[secretName]
	if !defined {
		return fmt.Errorf("Undefined certificate secret: " + secretName)
	}
	if dirtySecret, err = clusterRequest.reconcileCAs(secret, secretDef.caNames); err != nil {
		return err
	}
	if dirtySecret, err = clusterRequest.reconcileCerts(secret, secretDef.compCertDefs); err != nil {
		return err
	}
	if secretDef.sessionSecretName != nil {
		if dirtySecret, err = reconcileSessionSecret(secret, *secretDef.sessionSecretName); err != nil {
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
