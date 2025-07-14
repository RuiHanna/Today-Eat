package user

import (
	"backend/config"
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
)

// mock config.LoadServerConfig
var oldLoadServerConfig = config.LoadServerConfig

func mockLoadServerConfig(path string) (*config.ServerConfig, error) {
	return &config.ServerConfig{Domain: "http://test.com"}, nil
}

func TestUploadAvatarHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock db失败: %v", err)
	}
	defer db.Close()

	// 创建测试用的 config/server_config.json
	_ = os.MkdirAll("config", 0755)
	configContent := `{"domain":"http://test.com"}`
	configPath := filepath.Join("config", "server_config.json")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("写入测试配置文件失败: %v", err)
	}
	defer os.Remove(configPath) // 测试结束后删除

	r := gin.New()
	r.POST("/upload", UploadAvatarHandler(db))

	// 构造一张内存PNG图片
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var imgBuf bytes.Buffer
	if err := png.Encode(&imgBuf, img); err != nil {
		t.Fatalf("图片编码失败: %v", err)
	}

	// 构造multipart表单
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("user_id", "123")
	part, _ := writer.CreateFormFile("avatar", "avatar.png")
	part.Write(imgBuf.Bytes())
	writer.Close()

	// mock数据库
	mock.ExpectExec("UPDATE users SET avatar_url = \\? WHERE id = \\?").
		WithArgs("http://test.com/avatar/123.png", 123).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte("上传成功")) {
		t.Errorf("上传成功用例失败，返回: %s", w.Body.String())
	}
}
