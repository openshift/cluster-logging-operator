package common

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#GenerateSecretReaderScript", func() {

	It("should generate a valid script to read secrets", func() {
		exp, _ := embedFS.ReadFile("secret_reader_script.sh")
		secrets := helpers.Secrets{
			"mysecret": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "mysecret",
				},
				Data: map[string][]byte{
					"credentials": []byte("ggg"),
				},
			},
			"my-little-secret": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-little-secret",
				},
				Data: map[string][]byte{
					"password": []byte("ccc"),
				},
			},
			"other-secret": {
				ObjectMeta: metav1.ObjectMeta{
					Name: "other-secret",
				},
				Data: map[string][]byte{
					"some-token": []byte("ggg"),
				},
			},
		}

		script := GenerateSecretReaderScript(secrets)
		Expect(string(exp)).To(Equal(script))
	})
})
