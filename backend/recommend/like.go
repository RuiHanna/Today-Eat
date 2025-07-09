package recommend

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type LikeRequest struct {
	UserID int `json:"user_id"`
	DishID int `json:"dish_id"`
}

// 点赞接口
func LikeDish(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LikeRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 || req.DishID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		_, err := db.Exec("INSERT IGNORE INTO `like`(user_id, dish_id) VALUES (?, ?)", req.UserID, req.DishID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库写入失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "点赞成功"})
	}
}

// 取消点赞接口
func UnlikeDish(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LikeRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.UserID == 0 || req.DishID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		_, err := db.Exec("DELETE FROM `like` WHERE user_id = ? AND dish_id = ?", req.UserID, req.DishID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "取消点赞失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "已取消点赞"})
	}
}

// 获取用户收藏的菜品
func GetUserLikes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")

		rows, err := db.Query(`
			SELECT d.id, d.name, d.price, d.description, d.taste, d.score, d.image_url
			FROM `+"`like`"+` l
			JOIN dishes d ON l.dish_id = d.id
			WHERE l.user_id = ?
		`, userId)

		if err != nil {
			fmt.Println("查询失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询失败"})
			return
		}
		defer rows.Close()

		var favorites []map[string]interface{}
		for rows.Next() {
			var id int
			var name, desc, taste, imageURL string
			var price, score float64

			err := rows.Scan(&id, &name, &price, &desc, &taste, &score, &imageURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据解析错误"})
				return
			}

			favorites = append(favorites, gin.H{
				"id":          id,
				"name":        name,
				"price":       price,
				"description": desc,
				"taste":       taste,
				"score":       score,
				"image_url":   imageURL,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"code":      0,
			"favorites": favorites,
		})
	}
}
