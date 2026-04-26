package httpServer

import (
	"sync"

	"github.com/jxncyjq/stardust.mini/logs"
	"go.uber.org/zap"
)

// 客户端管理器
type ClientManager struct {
	mu         sync.RWMutex
	clients    map[string]IClient
	userIdMap  map[string]string // 用户ID到客户端的映射
	register   chan IClient
	unregister chan IClient
	broadcast  chan []byte
	stopChan   chan struct{}
	startChan  chan struct{}
	stopOnce   sync.Once
	logger     *zap.Logger
}

var Manager IClientManager

type IClientManager interface {
	Init()
	Start()
	ClientCount() int
	ClientKeepLive()
	KickClientByUserId(userId string)
	KickClientBySessionId(sessionId string)
	RegisterClient(client IClient)
	UnregisterClient(client IClient)
	BroadcastMessage(message []byte)
	Stop()
}

func NewClientManager(logger *zap.Logger) IClientManager {
	return &ClientManager{
		clients:    make(map[string]IClient),
		register:   make(chan IClient),
		unregister: make(chan IClient),
		broadcast:  make(chan []byte),
		stopChan:   make(chan struct{}),
		startChan:  make(chan struct{}),
		userIdMap:  make(map[string]string),
		logger:     logger,
	}
}

func (m *ClientManager) Init() {
	Manager = NewClientManager(logs.GetLogger("ClientManager"))
}

func (m *ClientManager) Start() {
	m.logger.Info("Starting client manager")
	close(m.startChan)
	for {
		select {
		//注册客户端
		case client := <-m.register:
			m.logger.Info("Registering new client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
			m.mu.Lock()
			m.clients[client.GetSessionID()] = client
			m.userIdMap[client.GetUserID()] = client.GetSessionID()
			m.mu.Unlock()
		// 删除客户端，并关健连接
		case client := <-m.unregister:
			m.logger.Info("Unregistering client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
			m.mu.Lock()
			if _, ok := m.clients[client.GetSessionID()]; ok {
				client.Close()
				delete(m.clients, client.GetSessionID())
			}
			m.mu.Unlock()
		// 所有消息广播
		case message := <-m.broadcast:
			m.mu.RLock()
			for _, client := range m.clients {
				m.logger.Debug("Broadcasting message to client", zap.String("sessionId", client.GetSessionID()))
				client.Send(message)
			}
			m.mu.RUnlock()
		case <-m.stopChan:
			m.logger.Debug("Stopping client manager")
			m.shutdownClients()
			m.logger.Info("Client manager stopped")
			return
		}
	}
}

func (m *ClientManager) shutdownClients() {
	m.mu.Lock()
	clients := make([]IClient, 0, len(m.clients))
	for sessionID, client := range m.clients {
		clients = append(clients, client)
		delete(m.clients, sessionID)
		delete(m.userIdMap, client.GetUserID())
	}
	m.mu.Unlock()

	for _, client := range clients {
		m.logger.Debug("Shutting down client", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
		if err := client.Close(); err != nil {
			m.logger.Warn("Failed to close client", zap.String("sessionId", client.GetSessionID()), logs.ErrorInfo(err))
		}
	}
	m.logger.Info("All clients have been kicked")
}

func (m *ClientManager) RegisterClient(client IClient) {
	select {
	case <-m.stopChan:
		m.logger.Debug("RegisterClient ignored: manager stopped", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
		return
	case m.register <- client:
	}
}

func (m *ClientManager) UnregisterClient(client IClient) {
	select {
	case <-m.stopChan:
		m.logger.Debug("UnregisterClient ignored: manager stopped", zap.String("sessionId", client.GetSessionID()), zap.String("userId", client.GetUserID()))
		return
	case m.unregister <- client:
	}
}

func (m *ClientManager) BroadcastMessage(message []byte) {
	select {
	case <-m.stopChan:
		m.logger.Debug("BroadcastMessage ignored: manager stopped")
		return
	case m.broadcast <- message:
	}
}

func (m *ClientManager) ClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// ClientKeepLive 定时发送心跳包
func (m *ClientManager) ClientKeepLive() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, client := range m.clients {
		client.Send([]byte("keepalive"))
	}
}

func (m *ClientManager) KickClientByUserId(userId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sessionId, ok := m.userIdMap[userId]; ok {
		if client, ok := m.clients[sessionId]; ok {
			client.Close()
			delete(m.clients, sessionId)
			delete(m.userIdMap, userId)
		}
	}
}

func (m *ClientManager) KickClientBySessionId(sessionId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if client, ok := m.clients[sessionId]; ok {
		client.Close()
		delete(m.clients, sessionId)
		delete(m.userIdMap, client.GetUserID())
	}
}

func (m *ClientManager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stopChan)
	})
	m.shutdownClients()
}
