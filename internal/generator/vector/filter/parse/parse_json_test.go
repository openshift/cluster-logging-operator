package parse

import (
	"strings"
	"testing"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func TestVRL(t *testing.T) {
	tests := []struct {
		name      string
		spec      *obs.ParseFilterSpec
		want      []string
		doNotWant []string
	}{
		{
			name:      "default",
			want:      []string{"._internal.structured = parsed"},
			doNotWant: []string{"merge!(parsed, .)", "parse_remove_message"},
		},
		{
			name:      "add to root",
			spec:      &obs.ParseFilterSpec{AddToRoot: true},
			want:      []string{"err == null && is_object(parsed)", "merge!(parsed, .)", "parse_remove_message = true"},
			doNotWant: []string{"._internal.structured = parsed"},
		},
		{
			name:      "add to root and preserve message",
			spec:      &obs.ParseFilterSpec{AddToRoot: true, PreserveMessage: true},
			want:      []string{"err == null && is_object(parsed)", "merge!(parsed, .)"},
			doNotWant: []string{"._internal.structured = parsed", "parse_remove_message"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			vrl, err := NewParseFilter(test.spec).VRL()
			if err != nil {
				t.Fatalf("VRL() returned an error: %v", err)
			}
			for _, want := range test.want {
				if !strings.Contains(vrl, want) {
					t.Errorf("VRL() does not contain %q:\n%s", want, vrl)
				}
			}
			for _, doNotWant := range test.doNotWant {
				if strings.Contains(vrl, doNotWant) {
					t.Errorf("VRL() unexpectedly contains %q:\n%s", doNotWant, vrl)
				}
			}
		})
	}
}
