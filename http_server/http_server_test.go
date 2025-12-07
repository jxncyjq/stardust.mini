package httpServer

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHttpServerHello(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/test"}`)
	srv, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create HttpServer: %v", err)
	}

	// 添加 hello 路由
	srv.Engine().GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	// 使用 httptest 测试
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	expected := `{"message":"hello"}`
	if w.Body.String() != expected {
		t.Errorf("Expected body %s, got %s", expected, w.Body.String())
	}
	t.Logf("Response: %s", w.Body.String())
}

func TestHttpServerHealthCheck(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/"}`)
	srv, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create HttpServer: %v", err)
	}

	srv.RegisterHealthCheck()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	t.Logf("Health check response: %s", w.Body.String())
}

func TestHttpServerHandler(t *testing.T) {
	config := []byte(`{"address":"127.0.0.1","port":0,"path":"/"}`)
	srv, err := NewHttpServer(config)
	if err != nil {
		t.Fatalf("Failed to create HttpServer: %v", err)
	}

	// 使用 Handler 泛型
	type HelloReq struct {
		Name string `json:"name"`
	}
	type HelloResp struct {
		Message string `json:"message"`
	}

	handler := NewHandler("hello", []string{"test"}, func(c *gin.Context, req HelloReq, resp HelloResp) error {
		c.JSON(http.StatusOK, gin.H{"message": "hello " + req.Name})
		return nil
	})

	srv.Get("hello", "", handler)

	req := httptest.NewRequest(http.MethodGet, "/api/hello?name=world", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	t.Logf("Handler response: %s", w.Body.String())
}
