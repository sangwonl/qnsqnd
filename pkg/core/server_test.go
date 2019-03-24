package core

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createServer() *gin.Engine {
	handlerCtx := HandlerContext{
		make(map[Topic]*TopicMessageInbox),
	}

	return setupRouter(&handlerCtx)
}

func TestPublishInNoTopicSubscribed(t *testing.T) {
	var jsonStr = []byte(`{"topic": "world", "payload": {}}`)
	req, _ := http.NewRequest("POST", "/publish", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	s := createServer()
	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	data := make(map[string]interface{})
	_ = json.Unmarshal([]byte(w.Body.String()), &data)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "invalid topic message", data["error"])
}
