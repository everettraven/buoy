package paneler

import (
	"strings"
)

func getDotNotationValue(item map[string]interface{}, dotPath string) interface{} {
	keys := strings.Split(dotPath, ".")
	var value interface{}
	curMap := item
	for _, key := range keys {
		val, ok := curMap[key]
		if !ok {
			return nil
		}
		newMap, ok := val.(map[string]interface{})
		if !ok {
			value = val
			break
		}
		curMap = newMap
	}

	return value
}
