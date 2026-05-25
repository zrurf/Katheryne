package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"ai-bot/internal/types"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type MessageHandler interface {
	HandleEvent(event *types.EventMessage)
	HandleNewMessage(data json.RawMessage)
	HandleMentionData(data json.RawMessage)
	HandleReply(msgID, content string)
}

type ClientConfig struct {
	TokenURL     string
	RefreshURL   string
	WSGatewayURL string
	ClientID     string
	ClientSecret string
}

type Client struct {
	config      ClientConfig
	oauth2      *OAuth2Manager
	conn        *websocket.Conn
	handler     MessageHandler
	closed      int32
	wg          sync.WaitGroup
	reconnectCh chan struct{}
	writeMu     sync.Mutex
}

func NewClient(cfg ClientConfig) *Client {
	return &Client{
		config:      cfg,
		oauth2:      NewOAuth2Manager(cfg.TokenURL, cfg.ClientID, cfg.ClientSecret),
		reconnectCh: make(chan struct{}, 1),
	}
}

func (c *Client) SetHandler(handler MessageHandler) {
	c.handler = handler
}

func (c *Client) Start() error {
	return c.connect()
}

func (c *Client) Stop() {
	atomic.StoreInt32(&c.closed, 1)
	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Wait()
}

func (c *Client) connect() error {
	token, err := c.oauth2.GetAccessToken()
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	conn, _, err := dialer.Dial(c.config.WSGatewayURL, header)
	if err != nil {
		return fmt.Errorf("dial ws gateway: %w", err)
	}
	c.conn = conn

	logx.Infof("AI Bot connected to WS gateway")

	c.wg.Add(2)
	go c.readPump()
	go c.writePump()

	return nil
}

func (c *Client) readPump() {
	defer c.wg.Done()
	defer func() {
		if atomic.LoadInt32(&c.closed) == 0 {
			c.ScheduleReconnect()
		}
	}()

	readTimeout := 90 * time.Second
	c.conn.SetReadLimit(65536)
	c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for atomic.LoadInt32(&c.closed) == 0 {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logx.Errorf("bot ws read error: %v", err)
			}
			return
		}

		c.conn.SetReadDeadline(time.Now().Add(readTimeout))

		var wsMsg types.WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			logx.Errorf("bot ws unmarshal error: %v", err)
			continue
		}

		switch wsMsg.Type {
		case "event":
			var event types.EventMessage
			if err := json.Unmarshal(wsMsg.Data, &event); err != nil {
				logx.Errorf("bot event unmarshal error: %v", err)
				continue
			}
			if c.handler != nil {
				c.handler.HandleEvent(&event)
			}

		case "new_message":
			// Wrap new_message data from ws-gateway into EventMessage format
			if c.handler != nil {
				c.handler.HandleNewMessage(wsMsg.Data)
			}

		case "mention":
			// Wrap mention data from ws-gateway into EventMessage format
			if c.handler != nil {
				c.handler.HandleMentionData(wsMsg.Data)
			}

		case "reply":
			if c.handler != nil {
				c.handler.HandleReply(wsMsg.Type, string(wsMsg.Data))
			}

		case "pong":
		}
	}
}

func (c *Client) writePump() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	writeTimeout := 10 * time.Second

	for atomic.LoadInt32(&c.closed) == 0 {
		select {
		case <-ticker.C:
			c.writeMu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			c.writeMu.Unlock()
			if err != nil {
				logx.Errorf("bot ws ping error: %v", err)
				return
			}
		}
	}
}

func (c *Client) SendMessage(data interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(data)
}

func (c *Client) ScheduleReconnect() {
	select {
	case c.reconnectCh <- struct{}{}:
		go func() {
			<-c.reconnectCh
			if atomic.LoadInt32(&c.closed) != 0 {
				return
			}

			backoff := 1 * time.Second
			maxBackoff := 60 * time.Second

			for atomic.LoadInt32(&c.closed) == 0 {
				logx.Infof("AI Bot reconnecting in %v...", backoff)
				time.Sleep(backoff)

				if err := c.connect(); err == nil {
					logx.Infof("AI Bot reconnected")
					return
				}

				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			}
		}()
	default:
	}
}
