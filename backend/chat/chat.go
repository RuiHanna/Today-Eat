package chat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSMessage struct {
	Message string `json:"message"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域连接，生产环境请限制
	},
}

func ChatWSHandler(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("WebSocket Upgrade error:", err)
			return
		}
		defer conn.Close()

		// 接收用户问题
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("接收消息失败"))
			return
		}

		var req WSMessage
		if err := json.Unmarshal(msgBytes, &req); err != nil || req.Message == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("格式错误"))
			return
		}

		// 发起 DeepSeek 流式请求
		err = callDeepSeekStream(apiKey, req.Message, func(content string) {
			conn.WriteMessage(websocket.TextMessage, []byte(content))
		})

		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("AI 调用失败: "+err.Error()))
		}
	}
}

func callDeepSeekStream(apiKey string, message string, onDelta func(string)) error {
	url := "https://api.deepseek.com/v1/chat/completions"

	bodyMap := map[string]interface{}{
		"model":  "deepseek-chat",
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个美食推荐助手，根据用户描述推荐菜品。"},
			{"role": "user", "content": message},
		},
		"temperature": 0.7,
	}

	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		// 处理每行：开头是 "data: "
		if bytes.HasPrefix(line, []byte("data: ")) {
			jsonPart := bytes.TrimPrefix(line, []byte("data: "))
			if bytes.Contains(jsonPart, []byte("[DONE]")) {
				break
			}
			var piece map[string]interface{}
			if err := json.Unmarshal(jsonPart, &piece); err == nil {
				if delta, ok := piece["choices"].([]interface{}); ok && len(delta) > 0 {
					if item, ok := delta[0].(map[string]interface{}); ok {
						if deltaVal, ok := item["delta"].(map[string]interface{}); ok {
							if content, ok := deltaVal["content"].(string); ok {
								onDelta(content)
							}
						}
					}
				}
			}
		}
	}
	return nil
}
