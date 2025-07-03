package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RegisterRequest 用户注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// RegisterHandler 注册处理函数
func RegisterHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

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
	}
}
