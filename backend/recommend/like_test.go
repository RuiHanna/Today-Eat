package recommend

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

func TestLikeDish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// mock db
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock db失败: %v", err)
	}
	defer db.Close()

	r := gin.New()
	r.POST("/like", LikeDish(db))

	// 1. 正常点赞
	mock.ExpectExec(`INSERT IGNORE INTO `+"`like`"+`\(user_id, dish_id\) VALUES \(\?, \?\)`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":1,"dish_id":2}`)
	req, _ := http.NewRequest("POST", "/like", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("点赞成功")) {
		t.Errorf("正常点赞失败，返回: %s", w.Body.String())
	}

	// 2. 参数错误
	w = httptest.NewRecorder()
	body = bytes.NewBufferString(`{"user_id":0,"dish_id":2}`)
	req, _ = http.NewRequest("POST", "/like", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("参数错误")) {
		t.Errorf("参数错误校验失败，返回: %s", w.Body.String())
	}

	// 3. 数据库写入失败
	mock.ExpectExec(`INSERT IGNORE INTO `+"`like`"+`\(user_id, dish_id\) VALUES \(\?, \?\)`).
		WithArgs(3, 4).
		WillReturnError(sql.ErrConnDone)

	w = httptest.NewRecorder()
	body = bytes.NewBufferString(`{"user_id":3,"dish_id":4}`)
	req, _ = http.NewRequest("POST", "/like", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError || !bytes.Contains(w.Body.Bytes(), []byte("数据库写入失败")) {
		t.Errorf("数据库写入失败未正确处理，返回: %s", w.Body.String())
	}
}

func TestUnlikeDish(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock db失败: %v", err)
	}
	defer db.Close()

	r := gin.New()
	r.POST("/unlike", UnlikeDish(db))

	// 1. 正常取消点赞
	mock.ExpectExec(`DELETE FROM `+"`like`"+` WHERE user_id = \? AND dish_id = \?`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	body := bytes.NewBufferString(`{"user_id":1,"dish_id":2}`)
	req, _ := http.NewRequest("POST", "/unlike", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("已取消点赞")) {
		t.Errorf("正常取消点赞失败，返回: %s", w.Body.String())
	}

	// 2. 参数错误
	w = httptest.NewRecorder()
	body = bytes.NewBufferString(`{"user_id":0,"dish_id":2}`)
	req, _ = http.NewRequest("POST", "/unlike", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest || !bytes.Contains(w.Body.Bytes(), []byte("参数错误")) {
		t.Errorf("参数错误校验失败，返回: %s", w.Body.String())
	}

	// 3. 数据库操作失败
	mock.ExpectExec(`DELETE FROM `+"`like`"+` WHERE user_id = \? AND dish_id = \?`).
		WithArgs(3, 4).
		WillReturnError(sql.ErrConnDone)

	w = httptest.NewRecorder()
	body = bytes.NewBufferString(`{"user_id":3,"dish_id":4}`)
	req, _ = http.NewRequest("POST", "/unlike", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError || !bytes.Contains(w.Body.Bytes(), []byte("取消点赞失败")) {
		t.Errorf("数据库操作失败未正确处理，返回: %s", w.Body.String())
	}
}

func TestGetUserLikes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock db失败: %v", err)
	}
	defer db.Close()

	r := gin.New()
	r.GET("/likes/:user_id", GetUserLikes(db))

	// 1. 正常返回
	rows := sqlmock.NewRows([]string{
		"id", "name", "price", "description", "taste", "score", "image_url",
	}).AddRow(1, "鱼香肉丝", 28.0, "经典川菜", "咸鲜微辣", 4.7, "http://img.com/1.jpg").
		AddRow(2, "宫保鸡丁", 32.0, "招牌菜", "微辣", 4.8, "http://img.com/2.jpg")

	mock.ExpectQuery("SELECT d.id, d.name, d.price, d.description, d.taste, d.score, d.image_url").
		WithArgs("123").
		WillReturnRows(rows)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/likes/123", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("favorites")) {
		t.Errorf("正常查询失败，返回: %s", w.Body.String())
	}

	// 2. 查询失败
	mock.ExpectQuery("SELECT d.id, d.name, d.price, d.description, d.taste, d.score, d.image_url").
		WithArgs("999").
		WillReturnError(sql.ErrConnDone)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/likes/999", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError || !bytes.Contains(w.Body.Bytes(), []byte("查询失败")) {
		t.Errorf("查询失败未正确处理，返回: %s", w.Body.String())
	}

	// 3. 数据解析失败（如 price 字段类型错误）
	badRows := sqlmock.NewRows([]string{
		"id", "name", "price", "description", "taste", "score", "image_url",
	}).AddRow(1, "鱼香肉丝", "not-a-float", "经典川菜", "咸鲜微辣", 4.7, "http://img.com/1.jpg")

	mock.ExpectQuery("SELECT d.id, d.name, d.price, d.description, d.taste, d.score, d.image_url").
		WithArgs("888").
		WillReturnRows(badRows)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/likes/888", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError || !bytes.Contains(w.Body.Bytes(), []byte("数据解析错误")) {
		t.Errorf("数据解析错误未正确处理，返回: %s", w.Body.String())
	}
}
