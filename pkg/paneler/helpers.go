package paneler

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func getDotNotationValue(item map[string]interface{}, dotPath string) (interface{}, error) {
	keys := strings.Split(dotPath, ".")
	val, exist, err := unstructured.NestedFieldNoCopy(item, keys...)
	if err != nil {
		return nil, fmt.Errorf("error fetching nested field %q: %w", dotPath, err)
	}
	if !exist {
		return nil, fmt.Errorf("nested field %q not found", dotPath)
	}
	return val, nil
}
