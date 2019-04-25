package k8shandler

import (
	"testing"
)

func TestSelectorsBothUndefined(t *testing.T) {

	commonSelector := map[string]string{}

	nodeSelector := map[string]string{}

	expected := map[string]string{}

	actual := mergeSelectors(nodeSelector, commonSelector)

	if !areSelectorsSame(actual, expected) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSelectorsCommonDefined(t *testing.T) {

	commonSelector := map[string]string{
		"common": "test",
	}

	nodeSelector := map[string]string{}

	expected := map[string]string{
		"common": "test",
	}

	actual := mergeSelectors(nodeSelector, commonSelector)

	if !areSelectorsSame(actual, expected) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSelectorsNodeDefined(t *testing.T) {

	commonSelector := map[string]string{}

	nodeSelector := map[string]string{
		"node": "test",
	}

	expected := map[string]string{
		"node": "test",
	}

	actual := mergeSelectors(nodeSelector, commonSelector)

	if !areSelectorsSame(actual, expected) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSelectorsCommonAndNodeDefined(t *testing.T) {

	commonSelector := map[string]string{
		"common": "test",
	}

	nodeSelector := map[string]string{
		"node": "test",
	}

	expected := map[string]string{
		"common": "test",
		"node":   "test",
	}

	actual := mergeSelectors(nodeSelector, commonSelector)

	if !areSelectorsSame(actual, expected) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}

func TestSelectorsCommonOverwritten(t *testing.T) {

	commonSelector := map[string]string{
		"common": "test",
		"node":   "test",
		"test":   "common",
	}

	nodeSelector := map[string]string{
		"common": "node",
		"test":   "node",
	}

	expected := map[string]string{
		"common": "node",
		"node":   "test",
		"test":   "node",
	}

	actual := mergeSelectors(nodeSelector, commonSelector)

	if !areSelectorsSame(actual, expected) {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}
