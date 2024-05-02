package utils

import (
	"errors"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

var (
	//It will map suffixes to make compatible with k8s "resource" package
	transformSuffix = map[string]string{
		"Ki": "Ki",
		"K":  "Ki",
		"k":  "Ki",
		"Mi": "Mi",
		"M":  "Mi",
		"m":  "Mi",
		"Gi": "Gi",
		"G":  "Gi",
		"g":  "Gi",
	}
)

// parseQuantityString is a fast scanner for quantity values.
// it extracts the numerical value and suffix from the quantity string
func parseQuantityString(str string) (value, suffix string, err error) {
	pos := 0
	end := len(str)
	// extract the value
Num:
	for i := pos; ; i++ {
		if i >= end {
			value = str[0:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			pos = i
			break Num
		}
	}
	value = str[0:pos]
	// grab the elements of the suffix
	suffixStart := pos
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		if !strings.ContainsAny(str[i:i+1], "eEinumgkKMGTP") {
			pos = i
			break
		}
	}
	if pos < end {
		switch str[pos] {
		case '-', '+':
			pos++
		}
	}
Suffix:
	for i := pos; ; i++ {
		if i >= end {
			suffix = str[suffixStart:end]
			return
		}
		switch str[i] {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		default:
			break Suffix
		}
	}
	// we encountered a non decimal in the Suffix loop, but the last character
	// was not a valid exponent
	err = errors.New("error occurred in parsing the quantity string")
	return value, suffix, err
}

// Serializes the values such as 750K, 8m, 1g to eqivalent base 10 value
func ParseQuantity(str string) (resource.Quantity, error) {
	value, suffix, err := parseQuantityString(str)
	if err != nil {
		return resource.Quantity{}, err
	}
	//newString is the value compatible with k8s "resource" package
	newString := value + transformSuffix[suffix]
	return resource.ParseQuantity(newString)
}
