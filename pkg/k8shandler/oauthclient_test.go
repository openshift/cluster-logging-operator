package k8shandler

import (
	"testing"
)

func TestAreOAuthClientsSameWhenEquivalent(t *testing.T) {
	one := NewOAuthClient("foo", "bar", "mysecret", []string{}, []string{})
	two := NewOAuthClient("otherfoo", "otherbar", "mysecret", []string{}, []string{})
	if !areOAuthClientsSame(one, two) {
		t.Errorf("Exp OAuthClients to be the same - left: %v, right: %v", one, two)
	}
}
func TestAreOAuthClientsSameWhenEquivalentScopesAndURIs(t *testing.T) {
	one := NewOAuthClient("foo", "bar", "mysecret", []string{"twouri", "oneuri"}, []string{"onescope", "twoscope"})
	two := NewOAuthClient("otherfoo", "otherbar", "mysecret", []string{"oneuri", "twouri"}, []string{"onescope", "twoscope"})
	if !areOAuthClientsSame(one, two) {
		t.Errorf("Exp OAuthClients to be the same - left: %v, right: %v", one, two)
	}
}
func TestAreOAuthClientsSameWhenDifferentScopesAndURIs(t *testing.T) {
	one := NewOAuthClient("foo", "bar", "mysecret", []string{"oneuri"}, []string{"onescope"})
	two := NewOAuthClient("otherfoo", "otherbar", "mysecret", []string{"twouri"}, []string{"twoscope"})
	if areOAuthClientsSame(one, two) {
		t.Errorf("Exp OAuthClients to be different - left: %v, right: %v", one, two)
	}
}
func TestAreOAuthClientsSameWhenDifferentSecrets(t *testing.T) {
	one := NewOAuthClient("foo", "bar", "mysecret", []string{"oneuri"}, []string{"onescope"})
	two := NewOAuthClient("otherfoo", "otherbar", "othersecret", []string{"oneuri"}, []string{"onescope"})
	if areOAuthClientsSame(one, two) {
		t.Errorf("Exp OAuthClients to be different - left: %v, right: %v", one, two)
	}
}
