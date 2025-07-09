package recommend

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// GetAllDishes 获取所有菜品(评分从高到低)
func GetAllDishes(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, price, description, taste, score, image_url, created_at FROM dishes ORDER BY score DESC")
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

// GetRandomDish 随机推荐菜品
func GetRandomDish(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Query("user_id") // 从请求查询参数获取用户ID
		if userIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少 user_id 参数"})
			return
		}

		var userID int
		_, err := fmt.Sscanf(userIDStr, "%d", &userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "user_id 参数无效"})
			return
		}

		// 查询5个随机菜品
		rows, err := db.Query(`
			SELECT id, name, price, description, taste, score, image_url, created_at 
			FROM dishes 
			ORDER BY RAND() 
			LIMIT 5
		`)
		if err != nil {
			fmt.Println("查询失败:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "message": "查询失败"})
			return
		}
		defer rows.Close()

		var dishes []gin.H
		var firstDishID int
		var firstDishLiked bool

		for rows.Next() {
			var d Dish
			err := rows.Scan(&d.ID, &d.Name, &d.Price, &d.Description, &d.Taste, &d.Score, &d.ImageURL, &d.CreatedAt)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据转换失败"})
				return
			}

			// 记录第一个菜品的ID
			if len(dishes) == 0 {
				firstDishID = d.ID
			}

			dishes = append(dishes, gin.H{
				"id":       d.ID,
				"name":     d.Name,
				"image":    d.ImageURL,
				"reason":   d.Description,
				"priceMin": int(d.Price * 0.9),
				"priceMax": int(d.Price * 1.2),
				"liked":    false, // 默认未点赞
			})
		}

		// 只查询第一个菜品的点赞情况
		if len(dishes) > 0 {
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM `like` WHERE user_id = ? AND dish_id = ?)", userID, firstDishID).Scan(&firstDishLiked)
			if err != nil {
				fmt.Println("查询点赞状态失败:", err)
				// 不返回错误，继续执行
			} else {
				// 更新第一个菜品的点赞状态
				dishes[0]["liked"] = firstDishLiked
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code":   0,
			"dishes": dishes,
		})
	}
}

// AddRecommendHistory 添加推荐历史
func AddRecommendHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID int `json:"user_id"`
			DishID int `json:"dish_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "请求参数错误"})
			return
		}

		_, err := db.Exec(`
			INSERT INTO recommend_history (user_id, dish_id) 
			VALUES (?, ?)`, req.UserID, req.DishID)

		if err != nil {
			fmt.Println("插入推荐历史失败：", err)
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库插入失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "记录成功"})
	}
}

// GetRecommendHistory 获取用户最近的推荐记录
func GetRecommendHistory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "缺少 user_id"})
			return
		}

		rows, err := db.Query(`
			SELECT d.id, d.name, d.price, d.description, d.taste, d.score, d.image_url
			FROM recommend_history rh
			JOIN dishes d ON rh.dish_id = d.id
			WHERE rh.user_id = ?
			ORDER BY rh.recommended_at DESC
		`, userID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库查询失败"})
			return
		}
		defer rows.Close()

		var history []Dish
		for rows.Next() {
			var d Dish
			err := rows.Scan(&d.ID, &d.Name, &d.Price, &d.Description, &d.Taste, &d.Score, &d.ImageURL)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 3, "message": "数据转换失败"})
				return
			}
			history = append(history, d)
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "history": history})
	}
}

type CustomRecord struct {
	UserID   int    `json:"user_id"`
	DishID   int    `json:"dish_id"`
	Taste    string `json:"taste"`
	Distance string `json:"distance"`
	Budget   int    `json:"budget"`
	Mood     string `json:"mood"`
	Weather  string `json:"weather"`
	Reason   string `json:"reason"`
}

// AddCustomRecordHandler 添加定制推荐记录
func AddCustomRecordHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CustomRecord
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数绑定失败"})
			return
		}

		stmt := `
			INSERT INTO custom_recommend_history 
			(user_id, dish_id, taste, distance, budget, mood, weather, reason, recommended_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`
		_, err := db.Exec(stmt,
			req.UserID, req.DishID, req.Taste, req.Distance,
			req.Budget, req.Mood, req.Weather, req.Reason, time.Now(),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "插入数据库失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "定制推荐记录已保存"})
	}
}

// GetDishDetailHandler 获取菜品详情
func GetDishDetailHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		dishIDStr := c.Query("id")
		if dishIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "缺少菜品 ID"})
			return
		}

		dishID, err := strconv.Atoi(dishIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 2, "message": "菜品 ID 无效"})
			return
		}

		var dish Dish

		err = db.QueryRow(`
			SELECT id, name, price, description, taste, score, image_url, created_at
			FROM dishes WHERE id = ?
		`, dishID).Scan(&dish.ID, &dish.Name, &dish.Price, &dish.Description, &dish.Taste, &dish.Score, &dish.ImageURL, &dish.CreatedAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "菜品不存在"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 4, "message": "数据库查询失败"})
			return
		}

		// 默认未点赞
		isLiked := false

		// 从 query 中获取 user_id
		userIDStr := c.Query("user_id")
		if userIDStr != "" {
			userID, err := strconv.Atoi(userIDStr)
			if err == nil {
				var count int
				err := db.QueryRow("SELECT COUNT(*) FROM `like` WHERE user_id = ? AND dish_id = ?", userID, dishID).Scan(&count)
				if err == nil && count > 0 {
					isLiked = true
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"id":          dish.ID,
				"name":        dish.Name,
				"priceMin":    int(dish.Price * 0.9),
				"priceMax":    int(dish.Price * 1.2),
				"taste":       dish.Taste,
				"score":       dish.Score,
				"description": dish.Description,
				"image":       dish.ImageURL,
				"liked":       isLiked,
			},
		})
	}
}
