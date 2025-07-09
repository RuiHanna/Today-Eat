package user

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RatingRequest 定义评分请求结构体
type RatingRequest struct {
	UserID int     `json:"user_id"`
	DishID int     `json:"dish_id"`
	Score  float64 `json:"score"`
}

// RateDishHandler 用户对菜品评分
func RateDishHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RatingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数绑定失败"})
			return
		}

		// 插入或更新评分
		stmt := `
			INSERT INTO dish_ratings (user_id, dish_id, score, rated_at)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE score = VALUES(score), rated_at = VALUES(rated_at)
		`
		_, err := db.Exec(stmt, req.UserID, req.DishID, req.Score, time.Now())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库操作失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "评分成功"})
	}
}
