package inputs

import (
	"testing"
)

func TestDeepCopy(t *testing.T) {
	a := Application{
		Namespaces: []string{"application-namespace"},
	}
	b := a.DeepCopy()
	if b == nil {
		t.Errorf("DeepCopy not success. b is nil")
		t.Fail()
	}
	if b.Namespaces == nil {
		t.Errorf("DeepCopy not success. b.Namespaces is nil")
		t.Fail()
	}
	if len(b.Namespaces) != 1 {
		t.Errorf("DeepCopy not success. len(b.Namespaces) is not 1")
		t.Fail()
	}
}

func TestDeepCopyInto(t *testing.T) {
	getA := func() *Application {
		return &Application{
			Namespaces: []string{"application-namespace"},
		}
	}
	getB := func() *Application {
		return &Application{}
	}
	tests := []struct {
		in      *Application
		out     *Application
		assertf func(*Application)
	}{
		{
			getA(),
			nil,
			func(b *Application) {
				if b != nil {
					t.Errorf("assert failed")
				}
			},
		},
		{
			nil,
			getB(),
			func(b *Application) {
				if b.Namespaces != nil {
					t.Errorf("assert failed")
				}
				if len(b.Namespaces) != 0 {
					t.Errorf("assert failed")
				}
			},
		},
		{
			getA(),
			getB(),
			func(b *Application) {
				if b.Namespaces == nil {
					t.Errorf("assert failed")
				}
				if len(b.Namespaces) != 1 {
					t.Errorf("assert failed")
				}
			},
		},
	}
	for _, test := range tests {
		test.in.DeepCopyInto(test.out)
		test.assertf(test.out)
	}
}
