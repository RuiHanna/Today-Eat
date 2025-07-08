package user

import (
	"database/sql"
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"

	"backend/config"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gin-gonic/gin"
)

func UploadAvatarHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.PostForm("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "用户ID错误"})
			return
		}

		file, fileHeader, err := c.Request.FormFile("avatar")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 2, "message": "上传失败"})
			return
		}
		defer file.Close()

		if fileHeader.Size > 2*1024*1024 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 9, "message": "图片太大，请上传小于2MB的图片"})
			return
		}

		// 解码图片为 image.Image
		img, _, err := image.Decode(file)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 3, "message": "图片解码失败（请上传 JPG/PNG 等标准图片）"})
			return
		}

		// 确保目录存在
		err = os.MkdirAll("data/avatar", os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 7, "message": "创建目录失败"})
			return
		}

		// 保存为 PNG 格式
		savePath := fmt.Sprintf("data/avatar/%d.png", userID)
		out, err := os.Create(savePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 4, "message": "保存头像失败"})
			return
		}
		defer out.Close()

		err = png.Encode(out, img) //图片写入文件夹
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 5, "message": "PNG 编码失败"})
			return
		}

		Cfg, err := config.LoadServerConfig("config/server_config.json")
		if err != nil {
			log.Fatalf("加载微信配置失败: %v", err)
		}

		// 构造头像访问 URL（替换为你的公网 IP 或域名）
		avatarURL := fmt.Sprintf("%s/avatar/%d.png", Cfg.Domain, userID)

		// 更新数据库头像地址
		_, err = db.Exec("UPDATE users SET avatar_url = ? WHERE id = ?", avatarURL, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 6, "message": "数据库更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "上传成功", "avatar_url": avatarURL})
	}
}
