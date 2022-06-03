package types

import (
	"regexp"
	"strconv"
	"strings"

	log "github.com/ViaQ/logerr/v2/log"
)

//OptionalInt allows passing an arbitrary int as well matching conditional values
//for message comparison (e.g. >=6).  A value of '' matches:
// - missing field
// - any value
type OptionalInt string

var emptyOptionalInt = NewOptionalInt("")

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
	logger := log.NewLogger("types-testing")
	logger.V(4).Info("Evaluating optionalInt", "value", oi)
	for i, name := range optionalIntRE.SubexpNames() {
		if name == "comparison" {
			comparison = parts[i]
		}
		if name == "value" {
			var err error
			value, err = strconv.Atoi(parts[i])
			if err != nil {
				log.NewLogger("").V(4).Error(err, "Unable to convert value into an OptionalInt. Expected a comparator and number (e.g. >12): defaulting to 0", "value", parts[i])
			}
		}

	}
	logger.V(4).Info("OptionalInt#getParts returning", "comparison", comparison, "value", value)
	return comparison, value
}

//IsSatisfiedBy returns true/false if the comparison needed is satisfied by other
func (oi OptionalInt) IsSatisfiedBy(other OptionalInt) bool {
	logger := log.NewLogger("types-testing")
	logger.V(4).Info("Comparing", "value", oi, "IsSatisfiedByArg", other)
	if oi == emptyOptionalInt {
		return true
	}
	var err error
	actValue := 0
	comparison, expValue := oi.getParts()
	actValue, err = strconv.Atoi(string(other))
	if err != nil {
		logger.V(4).Error(err, "Unable to parse actual value. returning false")
		return false
	}
	logger.V(4).Info("Expected", "comp", comparison, "value", expValue)
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
		return ""
	}
	return OptionalInt(value)
}
