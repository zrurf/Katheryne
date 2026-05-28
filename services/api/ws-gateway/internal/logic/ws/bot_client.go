package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"ws-gateway/internal/metrics"

	"conversation/conversationclient"
	"message/messageclient"
	"user/userclient"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type BotClient struct {
	botId    int64
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	lastPing int64
	closed   int32
	convIds  map[int64]bool
}

func NewBotClient(botId int64, hub *Hub, conn *websocket.Conn) *BotClient {
	return &BotClient{
		botId:    botId,
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		lastPing: time.Now().Unix(),
		convIds:  make(map[int64]bool),
	}
}

func (b *BotClient) ReadPump() {
	defer func() {
		b.hub.UnregisterBot(b)
		b.conn.Close()
	}()

	readTimeout := time.Duration(b.hub.config.ReadTimeout) * time.Second
	b.conn.SetReadLimit(b.hub.config.MaxMessageSize)
	b.conn.SetReadDeadline(time.Now().Add(readTimeout))
	b.conn.SetPongHandler(func(string) error {
		atomic.StoreInt64(&b.lastPing, time.Now().Unix())
		b.conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		_, message, err := b.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logx.Errorf("bot ws read error: bot_id=%d, err=%v", b.botId, err)
			}
			break
		}

		atomic.StoreInt64(&b.lastPing, time.Now().Unix())

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logx.Errorf("bot ws unmarshal error: bot_id=%d, err=%v", b.botId, err)
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("bot handleMessage panic recovered: bot_id=%d, err=%v", b.botId, r)
				}
			}()
			b.handleMessage(&msg)
		}()
	}
}

func (b *BotClient) WritePump() {
	ticker := time.NewTicker(time.Duration(b.hub.config.HeartbeatInterval) * time.Second)
	defer func() {
		ticker.Stop()
		b.conn.Close()
	}()

	writeTimeout := time.Duration(b.hub.config.WriteTimeout) * time.Second

	for {
		select {
		case message, ok := <-b.send:
			b.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				b.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := b.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logx.Errorf("bot ws write error: bot_id=%d, err=%v", b.botId, err)
				return
			}
		case <-ticker.C:
			b.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := b.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (b *BotClient) handleMessage(msg *WSMessage) {
	metrics.WsMessagesReceivedTotal.WithLabelValues("bot", msg.Type).Inc()

	switch msg.Type {
	case "ping":
		pong := MustNewWSMessage("pong", nil)
		data, _ := json.Marshal(pong)
		select {
		case b.send <- data:
		default:
		}

	case "send_message":
		var data BotSendMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid send_message data")
			return
		}
		b.handleSendMessage(&data)

	case "recall_message":
		var data BotRecallMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid recall_message data")
			return
		}
		b.handleRecallMessage(&data)

	case "get_message":
		var data BotGetMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid get_message data")
			return
		}
		b.handleGetMessage(&data)

	case "get_conversation":
		var data BotGetConversationData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid get_conversation data")
			return
		}
		b.handleGetConversation(&data)

	case "get_user":
		var data BotGetUserData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid get_user data")
			return
		}
		b.handleGetUser(&data)

	case "upload_file":
		var data BotUploadFileData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			b.sendError("invalid upload_file data")
			return
		}
		b.handleUploadFile(&data)

	default:
		b.sendError("unknown message type: " + msg.Type)
	}
}

func (b *BotClient) handleSendMessage(data *BotSendMessageData) {
	start := time.Now()
	defer func() {
		metrics.WsBotMessageLatency.WithLabelValues("send_message").Observe(time.Since(start).Seconds())
	}()

	metrics.WsMessagesSentTotal.WithLabelValues("bot", data.ContentType).Inc()
	metrics.WsBotMessagesTotal.WithLabelValues("send_message").Inc()

	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		b.sendError("invalid conv_id")
		return
	}
	var quoteMsgId int64
	if data.QuoteMsgId != "" {
		quoteMsgId, err = strconv.ParseInt(data.QuoteMsgId, 10, 64)
		if err != nil {
			quoteMsgId = 0
		}
	}

	msgType := strings.ToLower(data.MsgType)
	if msgType == "" {
		msgType = "text"
	}

	ctx := context.Background()
	resp, err := b.hub.config.MessageRpc.SendMessage(ctx, &messageclient.SendMessageReq{
		ConvId:      convId,
		Sender:      b.botId,
		Receiver:    0,
		Type:        msgType,
		Content:     data.Content,
		ContentType: data.ContentType,
		QuoteMsgId:  quoteMsgId,
		Extra:       data.Extra,
	})
	if err != nil {
		logx.Errorf("bot send message error: bot_id=%d, err=%v", b.botId, err)
		b.sendError("send message failed: " + err.Error())
		return
	}

	result := &BotSendMessageResp{
		MsgId:     strconv.FormatInt(resp.MsgId, 10),
		ConvId:    strconv.FormatInt(resp.ConvId, 10),
		CreatedAt: resp.CreatedAt,
	}
	wsMsg := MustNewWSMessage("send_message_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}

	b.hub.Broadcast(&BroadcastMsg{
		ConvId:       convId,
		Sender:       b.botId,
		Receiver:     0,
		MsgType:      msgType,
		Content:      data.Content,
		ContentType:  data.ContentType,
		SenderName:   data.SenderName,
		SenderAvatar: data.SenderAvatar,
		QuoteMsgId:   quoteMsgId,
		Extra:        data.Extra,
		MsgId:        resp.MsgId,
		CreatedAt:    resp.CreatedAt,
		ExcludeUid:   b.botId,
	})

	go func() {
		incrCtx, incrCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer incrCancel()
		membersResp, err := b.hub.config.ConversationRpc.GetConvMembers(incrCtx, &conversationclient.GetConvMembersReq{ConvId: convId})
		if err != nil {
			logx.Errorf("bot get conv members for unread error: convId=%d, err=%v", convId, err)
			return
		}
		var targetUids []int64
		for _, uid := range membersResp.Uids {
			if uid != b.botId {
				targetUids = append(targetUids, uid)
			}
		}
		if len(targetUids) > 0 {
			_, err = b.hub.config.ConversationRpc.IncrementUnread(incrCtx, &conversationclient.IncrementUnreadReq{
				ConvId: convId,
				Uids:   targetUids,
			})
			if err != nil {
				logx.Errorf("bot increment unread error: convId=%d, err=%v", convId, err)
			}
		}
	}()
}

func (b *BotClient) handleRecallMessage(data *BotRecallMessageData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		b.sendError("invalid conv_id")
		return
	}
	msgId, err := strconv.ParseInt(data.MsgId, 10, 64)
	if err != nil {
		b.sendError("invalid msg_id")
		return
	}

	ctx := context.Background()
	_, err = b.hub.config.MessageRpc.RecallMessage(ctx, &messageclient.RecallMessageReq{
		ConvId:   convId,
		MsgId:    msgId,
		Operator: b.botId,
	})
	if err != nil {
		logx.Errorf("bot recall message error: bot_id=%d, err=%v", b.botId, err)
		b.sendError("recall message failed: " + err.Error())
		return
	}

	result := &BotRecallMessageResp{Success: true}
	wsMsg := MustNewWSMessage("recall_message_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}

	push := &RecallMessagePush{
		ConvId: data.ConvId,
		MsgId:  data.MsgId,
		Uid:    strconv.FormatInt(b.botId, 10),
	}
	pushMsg := MustNewWSMessage("message_recalled", push)
	b.hub.PushToConv(convId, pushMsg, 0)
}

func (b *BotClient) handleGetMessage(data *BotGetMessageData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		b.sendError("invalid conv_id")
		return
	}
	msgId, err := strconv.ParseInt(data.MsgId, 10, 64)
	if err != nil {
		b.sendError("invalid msg_id")
		return
	}

	ctx := context.Background()
	resp, err := b.hub.config.MessageRpc.GetMessages(ctx, &messageclient.GetMessagesReq{
		ConvId:    convId,
		Cursor:    msgId,
		Limit:     1,
		Direction: "before",
	})
	if err != nil {
		logx.Errorf("bot get message error: bot_id=%d, err=%v", b.botId, err)
		b.sendError("get message failed: " + err.Error())
		return
	}

	result := &BotGetMessageResp{}
	if len(resp.List) > 0 {
		result.Msg = convertMsgItem(resp.List[0])
	} else {
		result.Error = "message not found"
	}

	wsMsg := MustNewWSMessage("get_message_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}
}

func (b *BotClient) handleGetConversation(data *BotGetConversationData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		b.sendError("invalid conv_id")
		return
	}

	ctx := context.Background()
	resp, err := b.hub.config.ConversationRpc.GetConversation(ctx, &conversationclient.GetConversationReq{
		ConvId: convId,
	})
	if err != nil {
		logx.Errorf("bot get conversation error: bot_id=%d, err=%v", b.botId, err)
		b.sendError("get conversation failed: " + err.Error())
		return
	}

	result := &BotGetConversationResp{
		ConvId: data.ConvId,
		Type:   resp.Type,
		Name:   resp.Name,
	}

	membersResp, err := b.hub.config.ConversationRpc.GetConvMembers(ctx, &conversationclient.GetConvMembersReq{
		ConvId: convId,
	})
	if err == nil {
		for _, uid := range membersResp.Uids {
			result.Members = append(result.Members, &BotConvMember{
				Uid: strconv.FormatInt(uid, 10),
			})
		}
	}

	wsMsg := MustNewWSMessage("get_conversation_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}
}

func (b *BotClient) handleGetUser(data *BotGetUserData) {
	uid, err := strconv.ParseInt(data.Uid, 10, 64)
	if err != nil {
		b.sendError("invalid uid")
		return
	}

	ctx := context.Background()
	resp, err := b.hub.config.UserRpc.GetUserByUID(ctx, &userclient.GetUserByUIDReq{
		Uid: uid,
	})
	if err != nil {
		logx.Errorf("bot get user error: bot_id=%d, err=%v", b.botId, err)
		b.sendError("get user failed: " + err.Error())
		return
	}

	result := &BotGetUserResp{
		Uid:    data.Uid,
		Name:   resp.User.Name,
		Avatar: resp.User.Avatar,
	}

	wsMsg := MustNewWSMessage("get_user_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}
}

func (b *BotClient) handleUploadFile(data *BotUploadFileData) {
	result := &BotUploadFileResp{
		FileId:  "",
		FileUrl: "",
		Error:   "file upload not implemented yet",
	}

	wsMsg := MustNewWSMessage("upload_file_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case b.send <- wsData:
	default:
	}
}

func (b *BotClient) sendError(message string) {
	push := &ErrorPush{
		Code:    400,
		Message: message,
	}
	wsMsg := MustNewWSMessage("error", push)
	data, _ := json.Marshal(wsMsg)
	select {
	case b.send <- data:
	default:
	}
}

func (b *BotClient) SendMessage(data []byte) {
	select {
	case b.send <- data:
	default:
	}
}

func (b *BotClient) Close() {
	if atomic.CompareAndSwapInt32(&b.closed, 0, 1) {
		b.conn.Close()
	}
}
