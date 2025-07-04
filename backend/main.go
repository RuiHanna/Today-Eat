package main

import (
	"backend/chat"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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

func main() {
	// 加载配置文件
	cfg, err := loadConfig("config.json")
	if err != nil {
		panic(err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 测试数据库连接
	if err := db.Ping(); err != nil {
		panic(fmt.Errorf("数据库无法连接: %v", err))
	}
	r := gin.Default()

	// 注册接口
	r.POST("/api/user/register", RegisterHandler(db)) // user.go 中的注册接口
	r.GET("/api/dishes", GetAllDishes(db))            // dishes.go 中的获取菜品接口
	r.POST("/api/chat", chat.ChatHandler(db))         // chat.go 中的聊天接口

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		panic(fmt.Errorf("服务器启动失败: %v", err))
	}
}
