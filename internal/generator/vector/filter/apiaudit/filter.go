package apiaudit

import (
	_ "embed"
	"encoding/json"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"net/http"
	"strings"
	"text/template"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func init() {
	utilruntime.Must(auditv1.AddToScheme(scheme.Scheme))
}

var DefaultOmitResponseCodes = []int{
	http.StatusNotFound,
	http.StatusConflict,
	http.StatusUnprocessableEntity,
	http.StatusTooManyRequests}

func PolicyToVRL(p *obs.KubeAPIAudit) (string, error) {
	if p == nil {
		p = &obs.KubeAPIAudit{} // Treat missing as empty.
	}
	if p.OmitResponseCodes == nil {
		p.OmitResponseCodes = &DefaultOmitResponseCodes
	}
	w := &strings.Builder{}
	err := policyVRLTemplate.Execute(w, p)
	return w.String(), err
}

//go:embed policy.vrl.tmpl
var policyVRLTemplateStr string

// This template compiles a Go k8s.io/apiserver/pkg/apis/audit/v1.Policy into a vector VRL remap program.
// The VRL program filters JSON-serialized k8s.io/apiserver/pkg/apis/audit/v1.Event objects according to the policy.
//
// NOTES:
// - Go templates and VRL both use "." to mean the current object, read carefully.
// - Go templates and VRL both use {{foo}} for substitution, this can be quoted as {{"{{"}}foo{{"}}"}} if you want VRL substitution.
var policyVRLTemplate = template.Must(template.New("policy VRL").Funcs(template.FuncMap{
	"json":         func(v any) (string, error) { b, err := json.Marshal(v); return string(b), err },
	"vsub":         func(v any) string { return fmt.Sprintf("{{%v}}", v) },
	"matchAny":     matchAny,
	"matchAnyPath": matchAnyPath,
}).Parse(policyVRLTemplateStr))
