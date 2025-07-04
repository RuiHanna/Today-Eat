package chat

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

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

func ChatHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.Message == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		// 提取关键词
		message := req.Message
		taste := detectTaste(message)

		if taste == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"reply": "暂时无法识别你想吃的口味 😥\n\n你可以试试说：\n- 我想吃辣的\n- 推荐几个清淡的菜\n- 有没有甜品推荐？",
			})
			return
		}

		// 查询菜品
		query := "SELECT name, description, taste FROM dishes WHERE taste LIKE ? LIMIT 3"
		rows, err := db.Query(query, "%"+taste+"%")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库查询失败"})
			return
		}
		defer rows.Close()

		var dishes []Dish
		for rows.Next() {
			var dish Dish
			if err := rows.Scan(&dish.Name, &dish.Description, &dish.Taste); err == nil {
				dishes = append(dishes, dish)
			}
		}

		if len(dishes) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"code":  0,
				"reply": fmt.Sprintf("我没找到 [%s] 口味的菜 😔，换一个试试吧～", taste),
			})
			return
		}

		// 生成 markdown 回复
		md := fmt.Sprintf("### 推荐的%s菜品\n\n", taste)
		for _, d := range dishes {
			md += fmt.Sprintf("- **%s**：%s\n", d.Name, d.Description)
		}
		md += "\n> 希望你喜欢这些推荐 🍽️"

		c.JSON(http.StatusOK, gin.H{
			"code":  0,
			"reply": md,
		})
	}
}

// 简单关键词匹配（可替换为NLP）
func detectTaste(message string) string {
	keywords := map[string]string{
		"辣":   "辣",
		"甜":   "甜",
		"酸":   "酸",
		"咸":   "咸",
		"清淡":  "清淡",
		"重口味": "辣",
		"麻":   "辣",
		"鲜":   "鲜",
	}

	for k, v := range keywords {
		if strings.Contains(message, k) {
			return v
		}
	}
	return ""
}
