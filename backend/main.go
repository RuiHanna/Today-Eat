package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

// 数据库配置
type Config struct {
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBName     string `json:"db_name"`
}

// 加载数据库
func loadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

func main() {
	cfg, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	r := gin.Default()

	r.POST("/api/user/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		// 检查用户名是否存在
		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&exists)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "数据库错误"})
			return
		}
		if exists > 0 {
			c.JSON(http.StatusOK, gin.H{"code": 3, "message": "用户名已存在"})
			return
		}

		// 加密密码
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 4, "message": "密码加密失败"})
			return
		}

		_, err = db.Exec("INSERT INTO users (username, password_hash, nickname) VALUES (?, ?, ?)", req.Username, string(hash), req.Nickname)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5, "message": "写入数据库失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "注册成功"})
	})

	r.Run(":8080")
}
