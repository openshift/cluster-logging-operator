package matchers

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	logger "github.com/ViaQ/logerr/log"
	"github.com/onsi/gomega/types"
)

type LogMatcher struct {
	expected interface{}
}

func FitLogFormatTemplate(expected interface{}) types.GomegaMatcher {
	return &LogMatcher{
		expected: expected,
	}
}

func (m *LogMatcher) Match(actual interface{}) (success bool, err error) {
	if reflect.TypeOf(m.expected) != reflect.TypeOf(actual) {
		return false, fmt.Errorf("matcher expects to compare same log types")
	}

	return CompareLog(m.expected, actual)
}

func (m *LogMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto fit \n\t%#v", actual, m.expected)
}

func (m *LogMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not fit \n\t%#v", actual, m.expected)
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func DeepFields(iface interface{}, namePrefix string) ([]reflect.Value, []string) {
	fields := make([]reflect.Value, 0)
	names := make([]string, 0)

	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)

	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		n := namePrefix + ifv.Type().Field(i).Name
		if !v.CanInterface() {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			typename := v.Type().Name()
			if typename != "Time" {
				moreFields, moreNames := DeepFields(v.Interface(), n+"_")
				fields = append(fields, moreFields...)
				names = append(names, moreNames...)
			} else {
				fields = append(fields, v)
				names = append(names, n)
			}
		case reflect.Ptr:
			if !isNil(v.Interface()) {
				elm := v.Elem()
				moreFields, moreNames := DeepFields(elm.Interface(), n+"_")
				fields = append(fields, moreFields...)
				names = append(names, moreNames...)
			}
		default:
			fields = append(fields, v)
			names = append(names, n)
		}
	}

	return fields, names
}

func compareLogLogic(name string, templateValue interface{}, value interface{}) bool {
	templateValueString := fmt.Sprintf("%v", templateValue)
	valueString := fmt.Sprintf("%v", value)

	if templateValueString == valueString { // Same value is ok
		logger.V(3).Info("CompareLogLogic: Same value for", "name", name, "value", valueString)
		return true
	}
	if templateValueString == "*" && valueString != "" { // Any value, not Nil is ok if template value is "*"
		logger.V(3).Info("CompareLogLogic: Any value for * ", "name", name, "value", valueString)
		return true
	}

	if templateValueString == "0001-01-01 00:00:00 +0000 UTC" && valueString != "" { // Any time value not Nil is ok if template value is empty time
		logger.V(3).Info("CompareLogLogic: Any value for 'empty time' ", "name", name, "value", valueString)
		return true
	}

	if strings.HasPrefix(templateValueString, "regex:") { // Using regex if starts with "/"
		match, _ := regexp.MatchString(templateValueString[6:], valueString)
		if match {
			logger.V(3).Info("CompareLogLogic: Fit regex ", "name", name, "value", valueString)
			return true
		}
	}

	logger.V(3).Info("CompareLogLogic: Mismatch !!!", "name", name, "templateValue", templateValueString, "value", valueString)
	return false
}

func CompareLog(template interface{}, log interface{}) (bool, error) {
	templateFieldValues, templateFieldNames := DeepFields(template, "")
	logFieldValues, logFieldNames := DeepFields(log, "")
	for i := range logFieldNames {
		logFieldValue := logFieldValues[i].Interface()
		logFieldName := logFieldNames[i]
		foundMatchFields := false
		for j := range templateFieldValues {
			templateFieldValue := templateFieldValues[j].Interface()
			templateFieldName := templateFieldNames[j]
			if templateFieldName == logFieldName {
				foundMatchFields = true
				logger.V(3).Info("CompareLog: comparing", "name", templateFieldName)
				if !isNil(templateFieldValue) { // Are we interested this field?
					if templateFieldValues[j].Kind() == reflect.Ptr { // Skip skeleton structure fields
						logger.V(3).Info("CompareLog: skipping skeleton", "name", templateFieldName)
						break
					}
					if compareLogLogic(templateFieldName, templateFieldValue, logFieldValue) {
						break
					}
					return false, nil
				} else {
					logger.V(3).Info("CompareLog: skipping not interesting field", "name", templateFieldName)
					break // If this is not an interesting field
				}
			}
		}
		if !foundMatchFields {
			logger.V(3).Info("CompareLog: skipping field, not found in template field", "name", logFieldName)
		}
	}

	return true, nil
}
