package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Dish 菜品类型
type Dish struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Taste       string  `json:"taste"`
	Score       float64 `json:"score"`
	ImageURL    string  `json:"image_url"`
	CreatedAt   string  `json:"created_at"`
}

// GetAllDishes 获取所有菜品
func GetAllDishes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, price, description, taste, score, image_url, created_at FROM dishes")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "数据库查询失败"})
			return
		}
		defer rows.Close()

		dishes := []Dish{}
		for rows.Next() {
			var d Dish
			err := rows.Scan(&d.ID, &d.Name, &d.Price, &d.Description, &d.Taste, &d.Score, &d.ImageURL, &d.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据转换失败"})
				return
			}
			dishes = append(dishes, d)
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": dishes})
	}
}
