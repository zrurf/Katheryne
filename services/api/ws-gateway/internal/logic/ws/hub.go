package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"ws-gateway/internal/metrics"

	"auth/authclient"
	"conversation/conversationclient"
	"message/messageclient"
	"social/socialclient"
	"user/userclient"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type HubConfig struct {
	Redis             *redis.Client
	HeartbeatInterval int64
	ReadTimeout       int64
	WriteTimeout      int64
	MaxMessageSize    int64
	AuthRpc           authclient.Auth
	UserRpc           userclient.User
	SocialRpc         socialclient.Social
	MessageRpc        messageclient.Message
	ConversationRpc   conversationclient.Conversation
}

type Hub struct {
	config HubConfig

	clients    map[int64]map[*Client]bool
	botClients map[int64]map[*BotClient]bool
	register   chan *Client
	unregister chan *Client
	botReg     chan *BotClient
	botUnreg   chan *BotClient
	broadcast  chan *BroadcastMsg
	mu         sync.RWMutex

	onlineUsers map[int64]bool
	onlineMu    sync.RWMutex

	ctx    context.Context
	cancel context.CancelFunc
}

type BroadcastMsg struct {
	ConvId      int64
	Sender      int64
	Receiver    int64
	MsgType     string
	Content     string
	ContentType string
	QuoteMsgId  int64
	Extra       string
	MsgId       int64
	CreatedAt   int64
	ExcludeUid  int64
}

func NewHub(cfg HubConfig) *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		config:      cfg,
		clients:     make(map[int64]map[*Client]bool),
		botClients:  make(map[int64]map[*BotClient]bool),
		register:    make(chan *Client, 1024),
		unregister:  make(chan *Client, 1024),
		botReg:      make(chan *BotClient, 1024),
		botUnreg:    make(chan *BotClient, 1024),
		broadcast:   make(chan *BroadcastMsg, 4096),
		onlineUsers: make(map[int64]bool),
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(time.Duration(h.config.HeartbeatInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.uid]; !ok {
				h.clients[client.uid] = make(map[*Client]bool)
			}
			h.clients[client.uid][client] = true
			connCount := len(h.clients[client.uid])
			h.mu.Unlock()

			h.onlineMu.Lock()
			h.onlineUsers[client.uid] = true
			h.onlineMu.Unlock()

			metrics.WsConnectionsTotal.WithLabelValues("client").Inc()
			metrics.WsConnectionsActive.Inc()
			metrics.OnlineUsers.Set(float64(len(h.onlineUsers)))

			h.broadcastOnlineStatus(client.uid, true)
			logx.Infof("client connected: uid=%d, active_connections=%d, total_online=%d", client.uid, connCount, len(h.onlineUsers))

		case client := <-h.unregister:
			h.mu.Lock()
			remainingConns := 0
			if clients, ok := h.clients[client.uid]; ok {
				delete(clients, client)
				remainingConns = len(clients)
				if remainingConns == 0 {
					delete(h.clients, client.uid)
					h.onlineMu.Lock()
					delete(h.onlineUsers, client.uid)
					h.onlineMu.Unlock()
					h.broadcastOnlineStatus(client.uid, false)
				}
			}
			h.mu.Unlock()
			close(client.send)

			metrics.WsConnectionsActive.Dec()
			metrics.OnlineUsers.Set(float64(len(h.onlineUsers)))
			logx.Infof("client disconnected: uid=%d, remaining_connections=%d, total_online=%d", client.uid, remainingConns, len(h.onlineUsers))

		case bot := <-h.botReg:
			h.mu.Lock()
			if _, ok := h.botClients[bot.botId]; !ok {
				h.botClients[bot.botId] = make(map[*BotClient]bool)
			}
			h.botClients[bot.botId][bot] = true
			h.mu.Unlock()

			metrics.WsConnectionsTotal.WithLabelValues("bot").Inc()
			metrics.WsBotConnectionsActive.Inc()
			logx.Infof("bot connected: bot_id=%d", bot.botId)

		case bot := <-h.botUnreg:
			h.mu.Lock()
			if bots, ok := h.botClients[bot.botId]; ok {
				delete(bots, bot)
				if len(bots) == 0 {
					delete(h.botClients, bot.botId)
				}
			}
			h.mu.Unlock()
			close(bot.send)

			metrics.WsBotConnectionsActive.Dec()
			logx.Infof("bot disconnected: bot_id=%d", bot.botId)

		case msg := <-h.broadcast:
			metrics.WsBroadcastTotal.Inc()
			h.handleBroadcast(msg)

		case <-ticker.C:
			h.checkHeartbeats()

		case <-h.ctx.Done():
			return
		}
	}
}

func (h *Hub) handleBroadcast(msg *BroadcastMsg) {
	senderName := ""
	senderAvatar := ""
	if msg.Sender > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		resp, err := h.config.UserRpc.GetUserByUID(ctx, &userclient.GetUserByUIDReq{Uid: msg.Sender})
		cancel()
		if err == nil && resp != nil && resp.User != nil {
			senderName = resp.User.Name
			senderAvatar = resp.User.Avatar
		}
	}

	push := &NewMessagePush{
		MsgId:        strconv.FormatInt(msg.MsgId, 10),
		ConvId:       strconv.FormatInt(msg.ConvId, 10),
		Sender:       strconv.FormatInt(msg.Sender, 10),
		SenderName:   senderName,
		SenderAvatar: senderAvatar,
		Type:         msg.MsgType,
		Content:      msg.Content,
		ContentType:  msg.ContentType,
		Extra:        msg.Extra,
		CreatedAt:    msg.CreatedAt,
	}
	if msg.Receiver > 0 {
		push.Receiver = strconv.FormatInt(msg.Receiver, 10)
	}
	if msg.QuoteMsgId > 0 {
		push.QuoteMsgId = strconv.FormatInt(msg.QuoteMsgId, 10)
	}

	wsMsg := MustNewWSMessage("new_message", push)
	data, err := json.Marshal(wsMsg)
	if err != nil {
		logx.Errorf("marshal broadcast message error: %v", err)
		return
	}

	if msg.Receiver > 0 {
		h.mu.RLock()
		if clients, ok := h.clients[msg.Receiver]; ok {
			for c := range clients {
				if c.uid == msg.ExcludeUid {
					continue
				}
				select {
				case c.send <- data:
				default:
				}
			}
		}
		if bots, ok := h.botClients[msg.Receiver]; ok {
			for b := range bots {
				select {
				case b.send <- data:
				default:
				}
			}
		}
		h.mu.RUnlock()
		return
	}

	convMembers := h.getConvMembers(msg.ConvId)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range convMembers {
		if uid == msg.ExcludeUid {
			continue
		}
		if clients, ok := h.clients[uid]; ok {
			for c := range clients {
				select {
				case c.send <- data:
				default:
				}
			}
		}
		if bots, ok := h.botClients[uid]; ok {
			for b := range bots {
				select {
				case b.send <- data:
				default:
				}
			}
		}
	}
}

func (h *Hub) getConvMembers(convId int64) []int64 {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := h.config.ConversationRpc.GetConvMembers(ctx, &conversationclient.GetConvMembersReq{ConvId: convId})
	if err != nil {
		logx.Errorf("get conv members error: convId=%d, err=%v", convId, err)
		return nil
	}
	return resp.Uids
}

func (h *Hub) broadcastOnlineStatus(uid int64, online bool) {
	status := "offline"
	if online {
		status = "online"
	}
	push := &OnlineStatusPush{
		Uid:    strconv.FormatInt(uid, 10),
		Status: status,
	}
	wsMsg := MustNewWSMessage("online_status", push)
	data, err := json.Marshal(wsMsg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, clients := range h.clients {
		for c := range clients {
			if c.uid == uid {
				continue
			}
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) checkHeartbeats() {
	now := time.Now().Unix()
	timeout := h.config.ReadTimeout

	h.mu.RLock()
	var deadClients []*Client
	for _, clients := range h.clients {
		for c := range clients {
			if now-atomic.LoadInt64(&c.lastPing) > timeout {
				deadClients = append(deadClients, c)
			}
		}
	}
	var deadBots []*BotClient
	for _, bots := range h.botClients {
		for b := range bots {
			if now-atomic.LoadInt64(&b.lastPing) > timeout {
				deadBots = append(deadBots, b)
			}
		}
	}
	h.mu.RUnlock()

	for _, c := range deadClients {
		metrics.WsHeartbeatFailures.Inc()
		c.Close()
	}
	for _, b := range deadBots {
		metrics.WsHeartbeatFailures.Inc()
		b.Close()
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

func (h *Hub) RegisterBot(bot *BotClient) {
	h.botReg <- bot
}

func (h *Hub) UnregisterBot(bot *BotClient) {
	h.botUnreg <- bot
}

func (h *Hub) Broadcast(msg *BroadcastMsg) {
	select {
	case h.broadcast <- msg:
	default:
		logx.Errorf("Broadcast channel FULL, dropping message: convId=%d, msgId=%d", msg.ConvId, msg.MsgId)
	}
}

func (h *Hub) PushToUser(uid int64, msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[uid]; ok {
		for c := range clients {
			select {
			case c.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) PushToUsers(uids []int64, msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range uids {
		if clients, ok := h.clients[uid]; ok {
			for c := range clients {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}
}

func (h *Hub) PushToConv(convId int64, msg *WSMessage, excludeUid int64) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	convMembers := h.getConvMembers(convId)

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, uid := range convMembers {
		if uid == excludeUid {
			continue
		}
		if clients, ok := h.clients[uid]; ok {
			for c := range clients {
				select {
				case c.send <- data:
				default:
				}
			}
		}
	}
}

func (h *Hub) PushToBot(botId int64, msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if bots, ok := h.botClients[botId]; ok {
		for b := range bots {
			select {
			case b.send <- data:
			default:
			}
		}
	}
}

func (h *Hub) IsUserOnline(uid int64) bool {
	h.onlineMu.RLock()
	defer h.onlineMu.RUnlock()
	return h.onlineUsers[uid]
}

func (h *Hub) GetOnlineUsers() []int64 {
	h.onlineMu.RLock()
	defer h.onlineMu.RUnlock()
	users := make([]int64, 0, len(h.onlineUsers))
	for uid := range h.onlineUsers {
		users = append(users, uid)
	}
	return users
}

func (h *Hub) Stop() {
	h.cancel()
}
