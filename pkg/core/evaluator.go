package core

import (
	"github.com/Knetic/govaluate"
	"github.com/gin-gonic/gin/json"
	"github.com/tidwall/gjson"
	"strings"
)

func customExpFuncs(jsonPayload interface{}) (map[string]govaluate.ExpressionFunction, error) {
	bytes, err := json.Marshal(jsonPayload)
	if err != nil {
		return nil, err
	}

	funcs := map[string]govaluate.ExpressionFunction{
		"F": func(args ...interface{}) (interface{}, error) {
			fieldExp := args[0].(string)
			if strings.HasPrefix(fieldExp, "@") {
				fieldExp = "\\" + fieldExp
			}
			result := gjson.GetBytes(bytes, fieldExp)
			return result.Value(), nil
		},
	}

	return funcs, nil
}

func evaluateExpression(jsonPayload interface{}, filter string) bool {
	functions, err := customExpFuncs(jsonPayload)
	if err != nil {
		return false
	}

	expression, _ := govaluate.NewEvaluableExpressionWithFunctions(filter, functions)
	result, _ := expression.Evaluate(nil)
	return result.(bool)
}
