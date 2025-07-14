package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWxLoginHandler_ParamError(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()

	router := gin.Default()
	router.POST("/api/user/wxlogin", WxLoginHandler(db, "appid", "appsecret"))

	body := `{"nickname":"小明"}`
	req, _ := http.NewRequest("POST", "/api/user/wxlogin", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(1), resp["code"])
}
