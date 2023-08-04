package sets

import (
	"github.com/golang-collections/collections/set"
	"sort"
)

type String struct {
	set.Set
}

func NewString(values ...string) *String {
	s := &String{
		*set.New(),
	}
	for _, v := range values {
		s.Insert(v)
	}
	return s
}

func (s *String) DeepCopyInto(in *String) {
	s.Do(func(entry interface{}) {
		in.Set.Insert(entry)
	})
}
func (s *String) Insert(values ...string) {
	for _, v := range values {
		s.Set.Insert(v)
	}
}

func (s *String) DeepCopy() *String {
	out := NewString()
	s.Do(func(entry interface{}) {
		out.Set.Insert(entry)
	})
	return out
}

func (s *String) List() []string {
	out := []string{}
	s.Do(func(entry interface{}) {
		s, _ := entry.(string)
		out = append(out, s)
	})
	sort.Strings(out)
	return out
}
