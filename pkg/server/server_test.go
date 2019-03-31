package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sangwonl/qnsqnd/pkg/core"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testServerInstance *gin.Engine
var once sync.Once

func getTestServerInstance() *gin.Engine {
	once.Do(func() {
		testServerInstance = setupRouter(&HandlerContext{
			core.GetPubSubManager(),
		})
	})
	return testServerInstance
}

type StreamRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *StreamRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *StreamRecorder) closeClient() {
	r.closeChannel <- true
}

func NewStreamRecorder() *StreamRecorder {
	return &StreamRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}

func reqPostJson(url string, jsonStr string) (*httptest.ResponseRecorder, map[string]interface{}) {
	s := getTestServerInstance()
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonStr)))
	req.Header.Set("Content-Type", "application/json")
	s.ServeHTTP(w, req)

	data := make(map[string]interface{})
	_ = json.Unmarshal([]byte(w.Body.String()), &data)

	return w, data
}

func reqGetStream(url string) *StreamRecorder {
	req, _ := http.NewRequest("GET", url, nil)
	w := NewStreamRecorder()
	s := getTestServerInstance()
	s.ServeHTTP(w, req)
	return w
}

func TestSubscribeWithEmptyTopic(t *testing.T) {
	jsonStr := `{"topic": ""}`
	w, res := reqPostJson("/subscribe", jsonStr)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "failed to subscribe", res["error"])
}

func TestPublishToNotSubscribedTopic(t *testing.T) {
	jsonStr := `{"topic": "world", "payload": {}}`
	w, res := reqPostJson("/publish", jsonStr)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "failed to publish topic message", res["error"])
}

func TestSubAndPubWithoutFilter(t *testing.T) {
	jsonStr := `{"topic": "news"}`
	w, res := reqPostJson("/subscribe", jsonStr)
	assert.Equal(t, 202, w.Code)

	subId, ok := res["subscriptionId"]
	assert.True(t, ok)
	assert.True(t, len(subId.(string)) == 36)

	NumMessages := 3
	TopicMessage := `{"topic": "news", "payload": {"hello": "world"}}`

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		sseUrl := fmt.Sprintf("/subscribe/%s/sse", subId)
		w := reqGetStream(sseUrl)

		body := w.Body.String()
		assert.True(t, len(body) > 0)
		assert.Equal(t, NumMessages, strings.Count(body, "event:message"))

		compact := strings.ReplaceAll(TopicMessage, " ", "")
		assert.Equal(t, NumMessages, strings.Count(body, compact))

		wg.Done()
	}()

	for i := 0; i < NumMessages; i++ {
		w, res = reqPostJson("/publish", TopicMessage)
		assert.Equal(t, 202, w.Code)
	}

	subCancelUrl := fmt.Sprintf("/subscribe/%s/cancel", subId)
	w, res = reqPostJson(subCancelUrl, "{}")
	assert.Equal(t, 202, w.Code)

	wg.Wait()
}
