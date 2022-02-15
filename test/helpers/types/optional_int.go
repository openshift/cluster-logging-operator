package types

import (
	"fmt"
	"github.com/ViaQ/logerr/log"
	"regexp"
	"strconv"
	"strings"
)

//OptionalInt allows passing an arbitrary int as well matching conditional values
//for message comparison (e.g. >=6)
type OptionalInt string

func (oi *OptionalInt) MarshalJSON() ([]byte, error) {
	return []byte(*oi), nil
}

func (oi *OptionalInt) UnmarshalJSON(data []byte) error {
	*oi = OptionalInt(string(data))
	return nil
}

func (oi OptionalInt) getParts() (string, int) {
	comparison := "="
	value := 0
	optionalIntRE := regexp.MustCompile(`(?P<comparison>[><=]{1,2})?(?P<value>\d*)`)
	parts := optionalIntRE.FindStringSubmatch(strings.TrimSpace(string(oi)))
	for i, name := range optionalIntRE.SubexpNames() {
		if name == "comparison" {
			comparison = parts[i]
		}
		if name == "value" {
			var err error
			value, err = strconv.Atoi(parts[i])
			if err != nil {
				log.V(4).Error(err, fmt.Sprintf("Unable to parse expected value %q into an OptionalInt. Expected a comparator and number (e.g. >12) and defaulting to 0", oi))
			}
		}

	}
	log.V(4).Info("OptionalInt#getParts returning", "comparison", comparison, "value", value)
	return comparison, value
}

//IsSatisfiedBy returns true/false if the comparison needed is satisfied by other
func (oi OptionalInt) IsSatisfiedBy(other OptionalInt) bool {
	var err error
	actValue := 0
	comparison, expValue := oi.getParts()
	actValue, err = strconv.Atoi(string(other))
	if err != nil {
		log.V(4).Error(err, fmt.Sprintf("Unable to parse the actual value %q into an OptionalInt. Expected a number (e.g. 12) and defaulting to 0", other))
	}
	log.V(4).Info("Expected", "comp", comparison, "value", expValue)
	switch comparison {
	case ">":
		return actValue > expValue
	case ">=":
		return actValue >= expValue
	case "<":
		return actValue < expValue
	case "<=":
		return actValue <= expValue
	default:
		return actValue == expValue
	}
}

func NewOptionalInt(value string) OptionalInt {
	if strings.TrimSpace(value) == "" {
		return OptionalInt("")
	}
	return OptionalInt(value)
}
