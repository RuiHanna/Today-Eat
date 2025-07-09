package recommend

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 定制推荐参数结构
type CustomRequest struct {
	UserID   int    `json:"user_id"`
	Taste    string `json:"taste"`
	Distance string `json:"distance"`
	Budget   int    `json:"budget"`
	Mood     string `json:"mood"`
	Weather  string `json:"weather"`
}

// CustomDishHandler 处理定制推荐请求
func CustomDishHandler(apiKey string, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CustomRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Println("❌ 请求绑定失败:", err)
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "请求参数错误"})
			return
		}

		// Step 1: 查询所有菜品
		rows, err := db.Query(`
			SELECT id, name, price, description, taste, score, image_url, created_at
			FROM dishes
		`)
		if err != nil {
			fmt.Println("❌ 数据库查询失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库查询失败"})
			return
		}
		defer rows.Close()

		var dishes []Dish
		for rows.Next() {
			var d Dish
			if err := rows.Scan(&d.ID, &d.Name, &d.Price, &d.Description, &d.Taste, &d.Score, &d.ImageURL, &d.CreatedAt); err == nil {
				dishes = append(dishes, d)
			} else {
				fmt.Println("❌ 读取菜品失败:", err)
			}
		}

		// Step 2: 构造 prompt
		prompt := buildPrompt(req, dishes)
		fmt.Println("📨 Prompt 提交给 AI:", prompt)

		// Step 3: 调用 DeepSeek（假设已封装好）
		selectedDish, reason, err := callDeepSeek(apiKey, prompt, dishes)
		if err != nil {
			fmt.Println("❌ AI推荐失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 3, "message": err.Error()})
			return
		}

		// Step 4: 构造响应
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"dish": gin.H{
				"id":       selectedDish.ID,
				"name":     selectedDish.Name,
				"image":    selectedDish.ImageURL,
				"reason":   reason,
				"priceMin": int(selectedDish.Price * 0.9),
				"priceMax": int(selectedDish.Price * 1.2),
				"liked":    false, // TODO: 可查 like 表
			},
		})
	}
}

// 构建 prompt
func buildPrompt(req CustomRequest, dishes []Dish) string {
	prompt := fmt.Sprintf(`你是一个美食推荐助手，用户的需求如下：
- 口味: %s
- 心情: %s
- 天气: %s
- 预算: %d 元以内

以下是候选菜品，请选择一个。
请按照以下格式输出：
推荐：<菜名>
理由：<推荐理由（不超过50字）>

`, req.Taste, req.Mood, req.Weather, req.Budget)

	for _, d := range dishes {
		prompt += fmt.Sprintf("菜品: %s｜价格: %.1f｜口味: %s｜描述: %s\n",
			d.Name, d.Price, d.Taste, d.Description)
	}
	return prompt
}

// 调用 DeepSeek API 推荐菜品
func callDeepSeek(apiKey string, prompt string, dishes []Dish) (Dish, string, error) {
	type DeepSeekRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}

	type DeepSeekResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	reqBody := DeepSeekRequest{
		Model: "deepseek-chat",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return Dish{}, "", err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return Dish{}, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ 请求 DeepSeek 失败:", err)
		return Dish{}, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Println("❌ DeepSeek 响应状态码:", resp.StatusCode)
		fmt.Println("❌ DeepSeek 响应内容:", buf.String())
		return Dish{}, "", fmt.Errorf("DeepSeek 接口返回错误状态 %d", resp.StatusCode)
	}

	var dsResp DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&dsResp); err != nil {
		fmt.Println("❌ DeepSeek 响应解析失败:", err)
		return Dish{}, "", err
	}
	if len(dsResp.Choices) == 0 {
		return Dish{}, "", errors.New("无推荐结果")
	}

	reply := dsResp.Choices[0].Message.Content
	fmt.Println("🤖 AI 回复:", reply)

	lines := strings.Split(reply, "\n")
	var name, reason string
	for _, line := range lines {
		if strings.HasPrefix(line, "推荐：") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "推荐："))
		} else if strings.HasPrefix(line, "理由：") {
			reason = strings.TrimSpace(strings.TrimPrefix(line, "理由："))
		}
	}

	if name == "" {
		return Dish{}, "", errors.New("AI输出中未提取到推荐菜名")
	}

	for _, d := range dishes {
		if d.Name == name {
			return d, reason, nil
		}
	}

	return Dish{}, "", fmt.Errorf("未在数据库中找到推荐菜品：%s", name)
}
