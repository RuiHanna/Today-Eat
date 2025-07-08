package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type WxLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type WxSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WxLoginHandler 微信登录处理函数
func WxLoginHandler(db *sql.DB, appID, appSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req WxLoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		openid, err := getOpenID(req.Code, appID, appSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": err.Error()})
			log.Printf("WxLoginHandler 出错: %v", err)
			return
		}

		var userID int
		var exists bool

		// 查询是否存在
		err = db.QueryRow("SELECT id FROM users WHERE openid = ?", openid).Scan(&userID)
		if err == sql.ErrNoRows {
			// 不存在，插入新用户
			res, err := db.Exec("INSERT INTO users (openid, nickname, avatar_url) VALUES (?, ?, ?)",
				openid, req.Nickname, req.AvatarURL)
			if err != nil {
				log.Printf("插入用户失败: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"code": 3, "message": "数据库写入失败"})
				return
			}
			lastID, _ := res.LastInsertId()
			userID = int(lastID)
		} else if err == nil {
			exists = true
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 4, "message": "数据库查询失败"})
			return
		}

		// 已存在则更新头像昵称（支持修改）
		if exists {
			_, err := db.Exec("UPDATE users SET nickname = ?, avatar_url = ? WHERE openid = ?",
				req.Nickname, req.AvatarURL, openid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 5, "message": "用户信息更新失败"})
				return
			}
		}

		// 写入 session
		session := sessions.Default(c)
		session.Set("user_id", userID)
		session.Save()

		c.JSON(http.StatusOK, gin.H{
			"code":     0,
			"message":  "登录成功",
			"user_id":  userID,
			"openid":   openid,
			"nickname": req.Nickname,
			"avatar":   req.AvatarURL,
		})
	}
}

// 获取openid
func getOpenID(code, appID, appSecret string) (string, error) {
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	// log.Printf("微信返回: %s", string(body))
	var wxResp WxSessionResponse
	err = json.Unmarshal(body, &wxResp)
	if err != nil {
		return "", err
	}

	if wxResp.ErrCode != 0 {
		return "", fmt.Errorf("微信返回错误: %s", wxResp.ErrMsg)
	}

	return wxResp.OpenID, nil
}

// UpdateNicknameHandler 更新昵称
func UpdateNicknameHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID   int    `json:"user_id"`
			Nickname string `json:"nickname"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "参数错误"})
			return
		}

		_, err := db.Exec("UPDATE users SET nickname = ? WHERE id = ?", req.Nickname, req.UserID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "更新失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "更新成功"})
	}
}

// LogoutHandler 登出处理函数
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "已登出"})
	}
}
