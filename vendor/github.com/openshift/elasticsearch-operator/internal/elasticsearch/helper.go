package elasticsearch

import (
	"strings"
)

func parseBool(path string, interfaceMap map[string]interface{}) bool {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedBool, ok := value.(bool); ok {
		return parsedBool
	} else {
		return false
	}
}

func parseString(path string, interfaceMap map[string]interface{}) string {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedString, ok := value.(string); ok {
		return parsedString
	} else {
		return ""
	}
}

func parseInt32(path string, interfaceMap map[string]interface{}) int32 {
	return int32(parseFloat64(path, interfaceMap))
}

func parseFloat64(path string, interfaceMap map[string]interface{}) float64 {
	value := walkInterfaceMap(path, interfaceMap)

	if parsedFloat, ok := value.(float64); ok {
		return parsedFloat
	} else {
		return float64(-1)
	}
}

func walkInterfaceMap(path string, interfaceMap map[string]interface{}) interface{} {
	current := interfaceMap
	keys := strings.Split(path, ".")
	keyCount := len(keys)

	for index, key := range keys {
		if current[key] != nil {
			if index+1 < keyCount {
				current = current[key].(map[string]interface{})
			} else {
				return current[key]
			}
		} else {
			return nil
		}
	}

	return nil
}
