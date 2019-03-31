package core

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func unmarshalJsonStr(jsonStr string) interface{} {
	var jsonData interface{}
	_ = json.Unmarshal([]byte(jsonStr), &jsonData)
	return jsonData
}

func TestJsonFieldEvaluation(t *testing.T) {
	jsonData := unmarshalJsonStr(`{
		"@metadata": {
			"beat": "filebeat",
			"topic": "qnsqnd",
			"type": "doc",
			"version": "1.2.3"
		},
		"@timestamp": "2019-02-01T14:45:30.761Z",
		"responseTime": 1004
	}`)

	assert.True(t, evaluateExpression(jsonData, `F("@metadata.version") == "1.2.3"`))
	assert.True(t, evaluateExpression(jsonData, `F("responseTime") > 1000`))
	assert.True(t, evaluateExpression(jsonData, `F("responseTime") - 4 == 1000`))
	assert.True(t, evaluateExpression(jsonData, `F("@metadata.topic") != "qnsqnd" ? false : true`))
}
