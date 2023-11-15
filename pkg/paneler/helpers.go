package paneler

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

func getDotNotationValue(item map[string]interface{}, dotPath string) (interface{}, error) {
	jsonBytes, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("error marshalling item to json: %w", err)
	}
	res := gjson.Get(string(jsonBytes), dotPath)
	if !res.Exists() {
		return nil, fmt.Errorf("nested field %q not found", dotPath)
	}
	return res.Value(), nil
}
