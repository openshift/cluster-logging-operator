package utils

import (
	"testing"
)

func TestAreMapsSameWhenBothAreEmpty(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{}
	if !AreMapsSame(one, two) {
		t.Error("Exp empty maps to evaluate to be equivalent")
	}
}
func TestAreMapsSameWhenOneIsEmptyAndTheOtherIsNot(t *testing.T) {
	one := map[string]string{}
	two := map[string]string{
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenEquivalent(t *testing.T) {
	one := map[string]string{
		"foo": "bar",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if !AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be equivalent - left: %v, right: %v", one, two)
	}
}
func TestAreMapsSameWhenDifferent(t *testing.T) {
	one := map[string]string{
		"foo": "456",
		"xyz": "123",
	}
	two := map[string]string{
		"xyz": "123",
		"foo": "bar",
	}
	if AreMapsSame(one, two) {
		t.Errorf("Exp maps to evaluate to be different - left: %v, right: %v", one, two)
	}
}
