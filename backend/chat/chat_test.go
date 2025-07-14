package chat

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

// mock callDeepSeekStream
var callDeepSeekStreamBak = callDeepSeekStream

func mockCallDeepSeekStream(apiKey string, message string, onDelta func(string)) error {
	onDelta("AI回复内容")
	return nil
}

func TestChatWSHandler(t *testing.T) {
	// 替换 callDeepSeekStream
	callDeepSeekStream = mockCallDeepSeekStream
	defer func() { callDeepSeekStream = callDeepSeekStreamBak }()

	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/ws", ChatWSHandler("fake-api-key"))

	server := httptest.NewServer(router)
	defer server.Close()

	// 将 http://127.0.0.1:xxxx 替换为 ws://127.0.0.1:xxxx
	wsURL := "ws" + server.URL[len("http"):] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// 发送消息
	msg := WSMessage{Message: "你好"}
	msgBytes, _ := json.Marshal(msg)
	err = conn.WriteMessage(websocket.TextMessage, msgBytes)
	assert.NoError(t, err)

	// 读取AI回复
	_, reply, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "AI回复内容", string(reply))
}
