package namespace

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/openshift/cluster-logging-operator/must-gather/internal/api"
	"github.com/openshift/cluster-logging-operator/must-gather/internal/client"
)

var _ = Describe("collectContainerLog", func() {
	var (
		tmpDir  string
		destDir api.Path
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "namespace-collector-test")
		Expect(err).ToNot(HaveOccurred())
		destDir = api.NewPath(tmpDir)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	// runAndCapture creates a collector with the given since options, calls collectContainerLog,
	// and returns the query parameters the Kubernetes client sent in the log request.
	runAndCapture := func(sinceSeconds *int64, sinceTime *metav1.Time) url.Values {
		captured := &url.Values{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*captured = r.URL.Query()
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "log line 1\nlog line 2\n")
		}))
		defer server.Close()

		k8sClient, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL})
		Expect(err).ToNot(HaveOccurred())

		n := &Collector{
			client:       &client.Client{KubernetesClient: k8sClient},
			logger:       api.NewLogger(io.Discard),
			sinceSeconds: sinceSeconds,
			sinceTime:    sinceTime,
		}
		Expect(n.collectContainerLog(context.Background(), "test-ns", "test-pod", "test-container", destDir, "current.log", false)).To(Succeed())
		return *captured
	}

	sincePtr := func(s int64) *int64 { return &s }
	sinceTimePtr := func(s string) *metav1.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return &metav1.Time{Time: t}
	}

	DescribeTable("honours since options in pod log requests",
		func(sinceSeconds *int64, sinceTime *metav1.Time, wantSinceSeconds, wantSinceTimeSubstr string) {
			q := runAndCapture(sinceSeconds, sinceTime)
			Expect(q.Get("sinceSeconds")).To(Equal(wantSinceSeconds))
			if wantSinceTimeSubstr != "" {
				Expect(q.Get("sinceTime")).To(ContainSubstring(wantSinceTimeSubstr))
			} else {
				Expect(q.Get("sinceTime")).To(BeEmpty())
			}
		},
		Entry("no since configuration omits both params",
			(*int64)(nil), (*metav1.Time)(nil), "", ""),
		Entry("SinceSeconds=300 sets sinceSeconds query param",
			sincePtr(300), (*metav1.Time)(nil), "300", ""),
		Entry("SinceTime sets sinceTime query param",
			(*int64)(nil), sinceTimePtr("2024-01-15T10:00:00Z"), "", "2024-01-15T10:00:00"),
	)
})
