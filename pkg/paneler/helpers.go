package paneler

import (
	"fmt"
	"strings"

	"github.com/everettraven/buoy/pkg/charm/models"
	"github.com/treilik/bubbleboxer"
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

func nodeForModelWrapper(name string, mw *models.ModelWrapper, bxr *bubbleboxer.Boxer) (bubbleboxer.Node, error) {
	node, err := bxr.CreateLeaf(name, mw)
	if err != nil {
		return bubbleboxer.Node{}, fmt.Errorf("creating leaf node: %w", err)
	}
	node.SizeFunc = func(node bubbleboxer.Node, widthOrHeight int) []int {
		return []int{mw.Height()}
	}
	return node, nil
}
