package main

import (
	"backend/chat"
	"backend/config"
	"backend/recommend"
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
	aiCfg, err := config.LoadAIConfig("config/ai_config.json")
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
	r.GET("/api/dishes", recommend.GetAllDishes(db))                                   // dishes.go 中的获取菜品接口
	r.GET("/api/chat/ws", chat.ChatWSHandler(aiCfg.APIKey))                            // chat.go 中的聊天接口
	r.POST("/api/user/wxlogin", user.WxLoginHandler(db, wxCfg.AppID, wxCfg.AppSecret)) //login.go 中的微信登录接口
	r.POST("/api/user/avatar", user.UploadAvatarHandler(db))                           //avatar.go 中的上传头像接口
	r.POST("/api/user/update_nickname", user.UpdateNicknameHandler(db))                //login.go 中的更新昵称接口
	r.GET("/api/dish/random", recommend.GetRandomDish(db))                             //randomRecom.go 中的随机推荐接口
	r.POST("/api/like/like", recommend.LikeDish(db))                                   //like.go 中的点赞接口
	r.POST("/api/like/unlike", recommend.UnlikeDish(db))                               //like.go 中的取消点赞接口
	r.GET("/api/user/:user_id/favorites", recommend.GetUserLikes(db))                  //like.go 中的获取用户收藏的菜品接口
	r.POST("/api/history/add", recommend.AddRecommendHistory(db))                      //dishes.go 中的添加推荐历史接口
	r.GET("/api/history", recommend.GetRecommendHistory(db))                           //dishes.go 中的获取推荐历史接口
	r.POST("/api/dish/custom", recommend.CustomDishHandler(aiCfg.APIKey, db))          //recommend.go 中的自定义推荐接口
	r.POST("/api/custom/add", recommend.AddCustomRecordHandler(db))                    //dishes.go 中的添加定制推荐记录接口
	r.GET("/api/user/info", user.GetUserInfoHandler(db))                               //login.go 中的获取用户完整信息接口
	r.GET("/api/dish/detail", recommend.GetDishDetailHandler(db))                      //dishes.go 中的获取菜品详情接口
	r.POST("/api/rating", user.RateDishHandler(db))                                    //rate.go 中的评分接口

	// 启动服务器
	if err := r.Run(":8080"); err != nil {
		panic(fmt.Errorf("服务器启动失败: %v", err))
	}
}
