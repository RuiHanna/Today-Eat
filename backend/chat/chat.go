package chat

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message string `json:"message"`
}

type Dish struct {
	Name        string
	Description string
	Taste       string
}

func ChatHandler(db *sql.DB, apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		reply, err := callDeepSeek(apiKey, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "AI 接口调用失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"reply": reply,
		})
	}
}

func callDeepSeek(apiKey string, message string) (string, error) {
	url := "https://api.deepseek.com/v1/chat/completions"

	bodyMap := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": "你是一个美食推荐助手，根据用户描述推荐菜品。"},
			{"role": "user", "content": message},
		},
		"temperature": 0.7,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("DeepSeek response:", string(respBody))

	// 提取 AI 回复内容
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("解析返回失败: %v", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("未找到回复内容: %s", string(respBody))
	}

	firstChoice, ok := choices[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("回复结构错误")
	}

	msg, ok := firstChoice["message"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("message 结构错误")
	}

	content, ok := msg["content"].(string)
	if !ok {
		return "", fmt.Errorf("回复内容无法读取")
	}

	return content, nil

}
