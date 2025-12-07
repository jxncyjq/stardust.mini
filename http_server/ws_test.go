package httpServer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jxncyjq/stardust.mini/codec"
	"github.com/jxncyjq/stardust.mini/logs"
)

func TestMain(m *testing.M) {
	logs.Init([]byte(`{"level":-1}`)) // -1 = debug level
	os.Exit(m.Run())
}

// TestMessageHandler 测试消息处理器
type TestMessageHandler struct{}

func (h *TestMessageHandler) HandlerMessage(message codec.IMessage) (string, error) {
	// 回显消息
	return `{"type":"echo","data":"` + message.GetType() + `"}`, nil
}

func TestWebSocketWithJsonCodec(t *testing.T) {
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	// 创建 WebSocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// 读取消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			t.Logf("Read error: %v", err)
			return
		}
		t.Logf("Server received: %s", msg)

		// 使用 codec 解码
		message, err := jsonCodec.Decode(msg)
		if err != nil {
			t.Logf("Decode error: %v", err)
			return
		}
		t.Logf("Decoded message type: %s", message.GetType())

		// 处理消息
		response, err := handler.HandlerMessage(message)
		if err != nil {
			t.Logf("Handler error: %v", err)
			return
		}

		// 发送响应
		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			t.Logf("Write error: %v", err)
		}
		t.Logf("Server sent: %s", response)
	}))
	defer server.Close()

	// 连接 WebSocket
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// 发送测试消息
	testMsg := `{"type":"ping","data":"hello"}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMsg))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	t.Logf("Client sent: %s", testMsg)

	// 读取响应
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	t.Logf("Client received: %s", response)

	// 验证响应
	respMsg, err := jsonCodec.Decode(response)
	if err != nil {
		t.Fatalf("Decode response failed: %v", err)
	}
	if respMsg.GetType() != "echo" {
		t.Errorf("Expected type 'echo', got '%s'", respMsg.GetType())
	}

	t.Log("WebSocket test with JsonCodec passed!")
}

func TestWebSocketClient(t *testing.T) {
	logger := logs.GetLogger("ws_client_test")
	jsonCodec := codec.NewJsonCodec()
	handler := &TestMessageHandler{}

	// 创建 mock ClientManager
	mockCM := &mockClientManager{}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// 创建 Client
		client := NewClient("user1", "session1", conn, jsonCodec, logger, context.Background(), handler, mockCM)
		t.Logf("Client created: userId=%s, sessionId=%s", client.GetUserID(), client.GetSessionID())

		// 直接处理一条消息
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		message, _ := jsonCodec.Decode(msg)
		response, _ := handler.HandlerMessage(message)
		conn.WriteMessage(websocket.TextMessage, []byte(response))
	}))
	defer server.Close()

	// 客户端连接
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer conn.Close()

	// 发送消息
	conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"test","data":"client_test"}`))

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	t.Logf("Response: %s", response)
	t.Log("WebSocket Client test passed!")
}

// mockClientManager 模拟 ClientManager
type mockClientManager struct{}

func (m *mockClientManager) Init()                              {}
func (m *mockClientManager) Start()                             {}
func (m *mockClientManager) ClientCount() int                   { return 0 }
func (m *mockClientManager) ClientKeepLive()                    {}
func (m *mockClientManager) KickClientByUserId(userId string)   {}
func (m *mockClientManager) KickClientBySessionId(sid string)   {}
func (m *mockClientManager) RegisterClient(client IClient)      {}
func (m *mockClientManager) UnregisterClient(client IClient)    {}
func (m *mockClientManager) BroadcastMessage(message []byte)    {}
func (m *mockClientManager) Stop() {}
