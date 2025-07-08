package main

import (
	"backend/chat"
	"backend/config"
	"backend/user"
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadDBConfig("config/db_config.json")
	if err != nil {
		panic(err)
	}
	wxCfg, err := config.LoadWxConfig("config/wx_config.json")
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

	r.Static("/avatar", "./data/avatar")

	store := cookie.NewStore([]byte("secret-key"))
	r.Use(sessions.Sessions("todayeat-session", store))

	// 注册接口
	r.GET("/api/dishes", GetAllDishes(db))                                             // dishes.go 中的获取菜品接口
	r.POST("/api/chat", chat.ChatHandler(db))                                          // chat.go 中的聊天接口
	r.POST("/api/user/wxlogin", user.WxLoginHandler(db, wxCfg.AppID, wxCfg.AppSecret)) //login.go 中的微信登录接口
	r.POST("/api/user/logout", user.LogoutHandler())                                   //login.go 中的登出接口
	r.POST("/api/user/avatar", user.UploadAvatarHandler(db))                           //avatar.go 中的上传头像接口
	r.POST("/api/user/update_nickname", user.UpdateNicknameHandler(db))                //login.go 中的更新昵称接口

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		panic(fmt.Errorf("服务器启动失败: %v", err))
	}
}
