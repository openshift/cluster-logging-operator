package k8shandler

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gomegatypes "github.com/onsi/gomega/types"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Reconciling", func() {
	_, filename, _, _ := runtime.Caller(0)
	scriptsDir := fmt.Sprintf("%s/../../scripts", filepath.Dir(filename))
	defer GinkgoRecover()

	_ = loggingv1.SchemeBuilder.AddToScheme(scheme.Scheme)

	var (
		cluster = &loggingv1.ClusterLogging{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "instance",
				Namespace: constants.OpenshiftNS,
			},
			Spec: loggingv1.ClusterLoggingSpec{
				ManagementState: loggingv1.ManagementStateManaged,
				Collection: &loggingv1.CollectionSpec{
					Logs: loggingv1.LogCollectionSpec{
						Type:        loggingv1.LogCollectionTypeFluentd,
						FluentdSpec: loggingv1.FluentdSpec{},
					},
				},
			},
		}
		masterCASecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.MasterCASecretName,
				Namespace: constants.OpenshiftNS,
			},
		}
		fluentdSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.FluentdName,
				Namespace: constants.OpenshiftNS,
			},
		}

		elasticsearchSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.ElasticsearchName,
				Namespace: constants.OpenshiftNS,
			},
		}

		kibanaSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.KibanaName,
				Namespace: constants.OpenshiftNS,
			},
		}

		kibanaProxySecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.KibanaProxyName,
				Namespace: constants.OpenshiftNS,
			},
		}

		curatorSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      constants.CuratorName,
				Namespace: constants.OpenshiftNS,
			},
		}
	)

	Describe("Certificates", func() {
		var (
			client         client.Client
			clusterRequest *ClusterLoggingRequest
		)

		BeforeEach(func() {
			client = fake.NewFakeClient(
				cluster,
				masterCASecret,
				fluentdSecret,
				elasticsearchSecret,
				kibanaSecret,
				kibanaProxySecret,
				curatorSecret,
			)
			clusterRequest = &ClusterLoggingRequest{
				client:  client,
				cluster: cluster,
			}

			os.Setenv("SCRIPTS_DIR", scriptsDir)
			Expect(clusterRequest.CreateOrUpdateCertificates()).To(Succeed())
		})

		It("should generate a self-signed CA", func() {
			secret := &corev1.Secret{}
			key := types.NamespacedName{Name: constants.MasterCASecretName, Namespace: constants.OpenshiftNS}
			Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

			Expect(secret).Should(SucceedParse(constants.MasterCASecretName))
			Expect(secret).Should(SucceedVerifyX509("masterca", "masterca", "openshift-cluster-logging-signer", nil, nil, nil))
		})

		Context("for log store", func() {
			It("should provide a X509 cert for `CN=elasticsearch`", func() {
				Expect(clusterRequest.createOrUpdateElasticsearchSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.ElasticsearchName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				san := []string{
					"elasticsearch-cluster",
					"elasticsearch.openshift-logging.svc",
					"elasticsearch.cluster.local",
				}
				Expect(secret).Should(SucceedParse(constants.ElasticsearchName))
				Expect(secret).Should(SucceedVerifyCert("admin-ca", "elasticsearch.crt", "elasticsearch", san))
			})

			It("should provide a X509 cert for `CN=logging-es`", func() {
				Expect(clusterRequest.createOrUpdateElasticsearchSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.ElasticsearchName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				san := []string{
					"elasticsearch.openshift-logging.svc",
					"elasticsearch.cluster.local",
				}
				Expect(secret).Should(SucceedParse(constants.ElasticsearchName))
				Expect(secret).Should(SucceedVerifyCert("admin-ca", "logging-es.crt", "elasticsearch", san))
			})

			It("should provide a X509 cert for `CN=system.admin`", func() {
				Expect(clusterRequest.createOrUpdateElasticsearchSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.ElasticsearchName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				Expect(secret).Should(SucceedParse(constants.ElasticsearchName))
				Expect(secret).Should(SucceedVerifyCert("admin-ca", "admin-cert", "system.admin", nil))
			})
		})

		Context("for collector", func() {
			It("should provide a X509 cert for `CN=system.logging.fluentd`", func() {
				Expect(clusterRequest.createOrUpdateFluentdSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.FluentdName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				Expect(secret).Should(SucceedParse(constants.FluentdName))
				Expect(secret).Should(SucceedVerifyCert("ca-bundle.crt", "tls.crt", "system.logging.fluentd", nil))
			})
		})

		Context("for visualization", func() {
			It("should provide a X509 cert for `CN=system.logging.kibana`", func() {
				Expect(clusterRequest.createOrUpdateKibanaSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.KibanaName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				Expect(secret).Should(SucceedParse(constants.KibanaName))
				Expect(secret).Should(SucceedVerifyCert("ca", "cert", "system.logging.kibana", nil))
			})

			It("should provide a X509 cert for `CN=kibanal-internal`", func() {
				Expect(clusterRequest.createOrUpdateKibanaSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.KibanaProxyName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				Expect(secret).Should(SucceedParse(constants.KibanaProxyName))
			})
		})

		Context("for curator", func() {
			It("should provide a X509 cert for `CN=system.logging.curator`", func() {
				Expect(clusterRequest.createOrUpdateCuratorSecret()).Should(Succeed())
				secret := &corev1.Secret{}
				key := types.NamespacedName{Name: constants.CuratorName, Namespace: constants.OpenshiftNS}
				Expect(client.Get(context.TODO(), key, secret)).Should(Succeed())

				Expect(secret).Should(SucceedParse(constants.CuratorName))
				Expect(secret).Should(SucceedVerifyCert("ca", "cert", "system.logging.curator", nil))
			})
		})
	})
})

func SucceedParse(secretName string) gomegatypes.GomegaMatcher {
	return &succeedParseMatcher{secretName: secretName}
}

type succeedParseMatcher struct {
	secretName string
}

func (matcher *succeedParseMatcher) Match(actual interface{}) (bool, error) {
	secret, ok := actual.(*corev1.Secret)
	if !ok || secret == nil {
		return false, fmt.Errorf("SuceedParse expects a non-nil *corev1.Secret")
	}

	if secret.Data == nil || len(secret.Data) == 0 {
		return false, fmt.Errorf("failed to read data from secret %q", matcher.secretName)
	}

	secretRef, ok := secretCertificates[matcher.secretName]
	if !ok {
		return false, fmt.Errorf("failed to load certificate reference for: %s", matcher.secretName)
	}

	for want := range secret.Data {
		found := false
		for k := range secretRef {
			if k == want {
				found = true
				break
			}
		}

		if !found {
			return false, fmt.Errorf("failed to find key %q for secret %q", want, matcher.secretName)
		}
	}

	return true, nil
}

func (matcher *succeedParseMatcher) FailureMessage(actual interface{}) (message string) {
	return ""
}

func (matcher *succeedParseMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ""
}

func SucceedVerifyCert(caKey, certKey, cn string, san []string) gomegatypes.GomegaMatcher {
	ou := []string{"OpenShift"}
	o := []string{"Logging"}
	return SucceedVerifyX509(caKey, certKey, cn, ou, o, san)
}

func SucceedVerifyX509(caKey, certKey, cn string, ou, o, san []string) gomegatypes.GomegaMatcher {
	return &succeedVerifyX509Matcher{
		caKey:   caKey,
		certKey: certKey,
		cn:      cn,
		ou:      ou,
		o:       o,
		san:     san,
	}
}

type succeedVerifyX509Matcher struct {
	caKey   string
	certKey string
	cn      string
	ou      []string
	o       []string
	san     []string
}

func (matcher *succeedVerifyX509Matcher) Match(actual interface{}) (bool, error) {
	secret, ok := actual.(*corev1.Secret)
	if !ok || secret == nil {
		return false, fmt.Errorf("SuceedVerifyX509 expects a non-nil *corev1.Secret")
	}

	if secret.Data == nil || len(secret.Data) == 0 {
		return false, fmt.Errorf("failed to read data from secret")
	}

	roots, cert, err := loadX509Cert(secret.Data[matcher.caKey], secret.Data[matcher.certKey])
	if err != nil {
		return false, fmt.Errorf("Failed to parse X509 certificate: %s", err.Error())
	}

	hosts := append(matcher.san, matcher.cn)
	for _, name := range hosts {
		opts := x509.VerifyOptions{
			DNSName: name,
			Roots:   roots,
		}
		if _, err := cert.Verify(opts); err != nil {
			return false, fmt.Errorf("failed to verify certificate: %s", err)
		}
	}

	if !reflect.DeepEqual(matcher.ou, cert.Subject.OrganizationalUnit) {
		return false, fmt.Errorf("failed to verify OU in certificate subject")
	}

	if !reflect.DeepEqual(matcher.o, cert.Subject.Organization) {
		return false, fmt.Errorf("failed to verify O in certificate subject")
	}

	return true, nil
}

func (matcher *succeedVerifyX509Matcher) FailureMessage(actual interface{}) (message string) {
	return ""
}

func (matcher *succeedVerifyX509Matcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ""
}

func loadX509Cert(ca, crt []byte) (*x509.CertPool, *x509.Certificate, error) {
	roots := x509.NewCertPool()

	ok := roots.AppendCertsFromPEM(ca)
	if !ok {
		return nil, nil, fmt.Errorf("failed to parse root certificate")
	}

	block, _ := pem.Decode(crt)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse certificate: " + err.Error())
	}
	return roots, cert, nil
}
