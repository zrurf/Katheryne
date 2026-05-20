package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"sync/atomic"
	"time"
	"ws-gateway/internal/metrics"

	"conversation/conversationclient"
	"message/messageclient"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type Client struct {
	uid      int64
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	lastPing int64
	closed   int32
}

func NewClient(uid int64, hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		uid:      uid,
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		lastPing: time.Now().Unix(),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		atomic.StoreInt32(&c.closed, 1)
		logx.Infof("ReadPump EXIT: uid=%d", c.uid)
		c.hub.Unregister(c)
		c.conn.Close()
	}()

	logx.Infof("ReadPump START: uid=%d", c.uid)

	readTimeout := time.Duration(c.hub.config.ReadTimeout) * time.Second
	c.conn.SetReadLimit(c.hub.config.MaxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	c.conn.SetPongHandler(func(string) error {
		atomic.StoreInt64(&c.lastPing, time.Now().Unix())
		c.conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				logx.Errorf("ws read error: uid=%d, err=%v", c.uid, err)
			}
			break
		}

		atomic.StoreInt64(&c.lastPing, time.Now().Unix())

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			logx.Errorf("ws unmarshal error: uid=%d, err=%v", c.uid, err)
			continue
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					logx.Errorf("handleMessage panic recovered: uid=%d, err=%v", c.uid, r)
				}
			}()
			c.handleMessage(&msg)
		}()
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(time.Duration(c.hub.config.HeartbeatInterval) * time.Second)
	defer func() {
		atomic.StoreInt32(&c.closed, 1)
		logx.Infof("WritePump EXIT: uid=%d", c.uid)
		ticker.Stop()
		c.conn.Close()
	}()

	logx.Infof("WritePump START: uid=%d", c.uid)

	writeTimeout := time.Duration(c.hub.config.WriteTimeout) * time.Second

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if !ok {
				logx.Infof("WritePump send channel closed: uid=%d", c.uid)
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			logx.Infof("WritePump writing message: uid=%d, len=%d", c.uid, len(message))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logx.Errorf("ws write error: uid=%d, err=%v", c.uid, err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logx.Errorf("ws ping write error: uid=%d, err=%v", c.uid, err)
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg *WSMessage) {
	metrics.WsMessagesReceivedTotal.WithLabelValues("client", msg.Type).Inc()

	logx.Infof("handleMessage received: uid=%d, type=%s, seq=%d, raw=%s", c.uid, msg.Type, msg.Seq, string(msg.Data))

	switch msg.Type {
	case "ping":
		pong := MustNewWSMessage("pong", nil)
		data, _ := json.Marshal(pong)
		select {
		case c.send <- data:
		default:
		}

	case "send_message":
		var data SendMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid send_message data")
			return
		}
		c.handleSendMessage(&data)

	case "read_receipt":
		var data ReadReceiptData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid read_receipt data")
			return
		}
		c.handleReadReceipt(&data)

	case "typing":
		var data TypingData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid typing data")
			return
		}
		c.handleTyping(&data)

	case "recall_message":
		var data RecallMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid recall_message data")
			return
		}
		c.handleRecallMessage(&data)

	case "edit_message":
		var data EditMessageData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid edit_message data")
			return
		}
		c.handleEditMessage(&data)

	case "search_messages":
		var data SearchMessagesData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid search_messages data")
			return
		}
		c.handleSearchMessages(&data)

	case "sync_offline":
		var data SyncOfflineData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid sync_offline data")
			return
		}
		c.handleSyncOffline(&data)

	case "get_messages":
		var data GetMessagesData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid get_messages data")
			return
		}
		c.handleGetMessages(&data)

	case "get_read_members":
		var data GetReadMembersData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			c.sendError(400, "invalid get_read_members data")
			return
		}
		c.handleGetReadMembers(&data)

	default:
		c.sendError(400, "unknown message type: "+msg.Type)
	}
}

func (c *Client) handleSendMessage(data *SendMessageData) {
	metrics.WsMessagesSentTotal.WithLabelValues("client", data.Type).Inc()

	logx.Infof("handleSendMessage START: uid=%d, convId=%s, type=%s, content=%s, tempId=%s", c.uid, data.ConvId, data.Type, data.Content, data.TempId)

	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		logx.Errorf("handleSendMessage invalid conv_id: uid=%d, convId=%s", c.uid, data.ConvId)
		c.sendError(400, "invalid conv_id")
		return
	}
	var receiver int64
	if data.Receiver != "" {
		receiver, err = strconv.ParseInt(data.Receiver, 10, 64)
		if err != nil {
			logx.Errorf("handleSendMessage invalid receiver: uid=%d, receiver=%s", c.uid, data.Receiver)
			c.sendError(400, "invalid receiver")
			return
		}
	}
	var quoteMsgId int64
	if data.QuoteMsgId != "" {
		quoteMsgId, err = strconv.ParseInt(data.QuoteMsgId, 10, 64)
		if err != nil {
			quoteMsgId = 0
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logx.Infof("handleSendMessage calling MessageRpc.SendMessage: uid=%d, convId=%d, sender=%d, receiver=%d, type=%s, contentType=%s", c.uid, convId, c.uid, receiver, data.Type, data.ContentType)
	resp, err := c.hub.config.MessageRpc.SendMessage(ctx, &messageclient.SendMessageReq{
		ConvId:      convId,
		Sender:      c.uid,
		Receiver:    receiver,
		Type:        data.Type,
		Content:     data.Content,
		ContentType: data.ContentType,
		QuoteMsgId:  quoteMsgId,
		Extra:       data.Extra,
	})
	if err != nil {
		logx.Errorf("handleSendMessage MessageRpc.SendMessage FAILED: uid=%d, convId=%d, err=%v", c.uid, convId, err)
		c.sendError(500, "send message failed: "+err.Error())
		return
	}
	logx.Infof("handleSendMessage MessageRpc.SendMessage SUCCESS: uid=%d, msgId=%d, convId=%d, createdAt=%d", c.uid, resp.MsgId, resp.ConvId, resp.CreatedAt)

	logx.Infof("handleSendMessage broadcasting: uid=%d, msgId=%d", c.uid, resp.MsgId)
	c.hub.Broadcast(&BroadcastMsg{
		ConvId:      convId,
		Sender:      c.uid,
		Receiver:    receiver,
		MsgType:     data.Type,
		Content:     data.Content,
		ContentType: data.ContentType,
		QuoteMsgId:  quoteMsgId,
		Extra:       data.Extra,
		MsgId:       resp.MsgId,
		CreatedAt:   resp.CreatedAt,
		ExcludeUid:  c.uid,
	})
	logx.Infof("handleSendMessage broadcast done: uid=%d, msgId=%d", c.uid, resp.MsgId)

	snippet := data.Content
	if len(snippet) > 100 {
		snippet = snippet[:100]
	}
	updateCtx, updateCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer updateCancel()
	_, err = c.hub.config.ConversationRpc.UpdateLastMessage(updateCtx, &conversationclient.UpdateLastMessageReq{
		ConvId:  convId,
		MsgId:   resp.MsgId,
		Snippet: snippet,
		Sender:  c.uid,
	})
	if err != nil {
		logx.Errorf("update last message error: convId=%d, err=%v", convId, err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("handleSendMessage incrUnread goroutine panic recovered: uid=%d, convId=%d, err=%v", c.uid, convId, r)
			}
		}()
		incrCtx, incrCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer incrCancel()
		membersResp, err := c.hub.config.ConversationRpc.GetConvMembers(incrCtx, &conversationclient.GetConvMembersReq{ConvId: convId})
		if err != nil {
			logx.Errorf("get conv members for unread error: convId=%d, err=%v", convId, err)
			return
		}
		var targetUids []int64
		for _, uid := range membersResp.Uids {
			if uid != c.uid {
				targetUids = append(targetUids, uid)
			}
		}
		if len(targetUids) > 0 {
			_, err = c.hub.config.ConversationRpc.IncrementUnread(incrCtx, &conversationclient.IncrementUnreadReq{
				ConvId: convId,
				Uids:   targetUids,
			})
			if err != nil {
				logx.Errorf("increment unread error: convId=%d, err=%v", convId, err)
			}
		}
	}()

	ack := MustNewWSMessage("send_message_ack", map[string]interface{}{
		"msg_id":     strconv.FormatInt(resp.MsgId, 10),
		"conv_id":    strconv.FormatInt(resp.ConvId, 10),
		"created_at": resp.CreatedAt,
		"temp_id":    data.TempId,
	})
	ackData, _ := json.Marshal(ack)
	if atomic.LoadInt32(&c.closed) == 1 {
		logx.Errorf("handleSendMessage ack DROPPED (client closed): uid=%d, msgId=%d", c.uid, resp.MsgId)
		return
	}
	select {
	case c.send <- ackData:
		logx.Infof("handleSendMessage ack sent: uid=%d, msgId=%d, tempId=%s", c.uid, resp.MsgId, data.TempId)
	default:
		logx.Errorf("handleSendMessage ack DROPPED (channel full): uid=%d, msgId=%d", c.uid, resp.MsgId)
	}
}

func (c *Client) handleReadReceipt(data *ReadReceiptData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		return
	}
	lastReadMsgId, err := strconv.ParseInt(data.LastReadMsgId, 10, 64)
	if err != nil {
		return
	}
	var startMsgId, endMsgId int64
	if data.StartMsgId != "" {
		startMsgId, _ = strconv.ParseInt(data.StartMsgId, 10, 64)
	}
	if data.EndMsgId != "" {
		endMsgId, _ = strconv.ParseInt(data.EndMsgId, 10, 64)
	}

	ctx := context.Background()
	_, err = c.hub.config.MessageRpc.SubmitRead(ctx, &messageclient.SubmitReadReq{
		ConvId:        convId,
		Uid:           c.uid,
		LastReadMsgId: lastReadMsgId,
		StartMsgId:    startMsgId,
		EndMsgId:      endMsgId,
	})
	if err != nil {
		logx.Errorf("submit read error: uid=%d, err=%v", c.uid, err)
		return
	}

	push := &ReadReceiptPush{
		ConvId:        data.ConvId,
		Uid:           strconv.FormatInt(c.uid, 10),
		LastReadMsgId: data.LastReadMsgId,
		StartMsgId:    data.StartMsgId,
		EndMsgId:      data.EndMsgId,
	}
	wsMsg := MustNewWSMessage("read_receipt", push)
	c.hub.PushToConv(convId, wsMsg, c.uid)
}

func (c *Client) handleTyping(data *TypingData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		return
	}

	go func() {
		push := &TypingPush{
			ConvId: data.ConvId,
			Uid:    strconv.FormatInt(c.uid, 10),
			Status: data.Status,
		}
		wsMsg := MustNewWSMessage("typing", push)
		c.hub.PushToConv(convId, wsMsg, c.uid)
	}()
}

func (c *Client) handleRecallMessage(data *RecallMessageData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid conv_id")
		return
	}
	msgId, err := strconv.ParseInt(data.MsgId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid msg_id")
		return
	}

	ctx := context.Background()
	_, err = c.hub.config.MessageRpc.RecallMessage(ctx, &messageclient.RecallMessageReq{
		ConvId:   convId,
		MsgId:    msgId,
		Operator: c.uid,
	})
	if err != nil {
		logx.Errorf("recall message error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "recall message failed: "+err.Error())
		return
	}

	push := &RecallMessagePush{
		ConvId: data.ConvId,
		MsgId:  data.MsgId,
		Uid:    strconv.FormatInt(c.uid, 10),
	}
	wsMsg := MustNewWSMessage("message_recalled", push)
	c.hub.PushToConv(convId, wsMsg, 0)
}

func (c *Client) handleEditMessage(data *EditMessageData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid conv_id")
		return
	}
	msgId, err := strconv.ParseInt(data.MsgId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid msg_id")
		return
	}

	ctx := context.Background()
	_, err = c.hub.config.MessageRpc.EditMessage(ctx, &messageclient.EditMessageReq{
		ConvId:  convId,
		MsgId:   msgId,
		Content: data.Content,
		Editor:  c.uid,
	})
	if err != nil {
		logx.Errorf("edit message error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "edit message failed: "+err.Error())
		return
	}

	push := &EditMessagePush{
		ConvId:  data.ConvId,
		MsgId:   data.MsgId,
		Content: data.Content,
		Uid:     strconv.FormatInt(c.uid, 10),
	}
	wsMsg := MustNewWSMessage("message_edited", push)
	c.hub.PushToConv(convId, wsMsg, 0)
}

func (c *Client) handleSearchMessages(data *SearchMessagesData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		convId = 0
	}
	sender, err := strconv.ParseInt(data.Sender, 10, 64)
	if err != nil {
		sender = 0
	}

	ctx := context.Background()
	resp, err := c.hub.config.MessageRpc.SearchMessages(ctx, &messageclient.SearchMessagesReq{
		Keyword:   data.Keyword,
		ConvId:    convId,
		Sender:    sender,
		StartTime: data.StartTime,
		EndTime:   data.EndTime,
		Page:      data.Page,
		Size:      data.Size,
	})
	if err != nil {
		logx.Errorf("search messages error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "search messages failed: "+err.Error())
		return
	}

	result := &SearchMessagesResp{
		Total: resp.Total,
	}
	for _, m := range resp.List {
		result.List = append(result.List, convertMsgItem(m))
	}

	wsMsg := MustNewWSMessage("search_messages_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case c.send <- wsData:
	default:
	}
}

func (c *Client) handleSyncOffline(data *SyncOfflineData) {
	ctx := context.Background()

	convResp, err := c.hub.config.ConversationRpc.GetConversations(ctx, &conversationclient.GetConversationsReq{Uid: c.uid})
	var convIds []int64
	if err != nil {
		logx.Errorf("get conversations for sync offline error: uid=%d, err=%v", c.uid, err)
	} else {
		for _, conv := range convResp.List {
			convIds = append(convIds, conv.ConvId)
		}
	}

	resp, err := c.hub.config.MessageRpc.SyncOfflineMsgs(ctx, &messageclient.SyncOfflineMsgsReq{
		Uid:     c.uid,
		Limit:   data.Limit,
		ConvIds: convIds,
	})
	if err != nil {
		logx.Errorf("sync offline error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "sync offline failed: "+err.Error())
		return
	}

	result := &SyncOfflineResp{}
	for _, m := range resp.Messages {
		result.List = append(result.List, convertMsgItem(m))
	}

	wsMsg := MustNewWSMessage("sync_offline_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case c.send <- wsData:
	default:
	}
}

func (c *Client) handleGetMessages(data *GetMessagesData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid conv_id")
		return
	}
	cursor, err := strconv.ParseInt(data.Cursor, 10, 64)
	if err != nil {
		cursor = 0
	}

	ctx := context.Background()
	resp, err := c.hub.config.MessageRpc.GetMessages(ctx, &messageclient.GetMessagesReq{
		ConvId:    convId,
		Cursor:    cursor,
		Limit:     data.Limit,
		Direction: data.Direction,
	})
	if err != nil {
		logx.Errorf("get messages error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "get messages failed: "+err.Error())
		return
	}

	result := &GetMessagesResp{
		HasMore: resp.HasMore,
	}
	for _, m := range resp.List {
		result.List = append(result.List, convertMsgItem(m))
	}

	wsMsg := MustNewWSMessage("get_messages_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case c.send <- wsData:
	default:
	}
}

func (c *Client) handleGetReadMembers(data *GetReadMembersData) {
	convId, err := strconv.ParseInt(data.ConvId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid conv_id")
		return
	}
	msgId, err := strconv.ParseInt(data.MsgId, 10, 64)
	if err != nil {
		c.sendError(400, "invalid msg_id")
		return
	}

	ctx := context.Background()
	resp, err := c.hub.config.MessageRpc.GetReadMembers(ctx, &messageclient.GetReadMembersReq{
		ConvId: convId,
		MsgId:  msgId,
	})
	if err != nil {
		logx.Errorf("get read members error: uid=%d, err=%v", c.uid, err)
		c.sendError(500, "get read members failed: "+err.Error())
		return
	}

	result := &GetReadMembersResp{
		Total: resp.Total,
	}
	for _, m := range resp.List {
		result.List = append(result.List, &ReadMemberItem{
			Uid:    strconv.FormatInt(m.Uid, 10),
			Name:   m.Name,
			Avatar: m.Avatar,
			ReadAt: m.ReadAt,
		})
	}

	wsMsg := MustNewWSMessage("get_read_members_resp", result)
	wsData, _ := json.Marshal(wsMsg)
	select {
	case c.send <- wsData:
	default:
	}
}

func (c *Client) sendError(code int64, message string) {
	push := &ErrorPush{
		Code:    code,
		Message: message,
	}
	wsMsg := MustNewWSMessage("error", push)
	data, _ := json.Marshal(wsMsg)
	select {
	case c.send <- data:
	default:
		logx.Errorf("sendError DROPPED (channel full): uid=%d, code=%d, message=%s", c.uid, code, message)
	}
}

func (c *Client) SendMessage(data []byte) {
	select {
	case c.send <- data:
	default:
	}
}

func (c *Client) Close() {
	if atomic.CompareAndSwapInt32(&c.closed, 0, 1) {
		c.conn.Close()
	}
}

func convertMsgItem(m *messageclient.MsgItem) *MsgItem {
	item := &MsgItem{
		Id:          strconv.FormatInt(m.Id, 10),
		ConvId:      strconv.FormatInt(m.ConvId, 10),
		Sender:      strconv.FormatInt(m.Sender, 10),
		Receiver:    strconv.FormatInt(m.Receiver, 10),
		Type:        m.Type,
		Content:     m.Content,
		ContentType: m.ContentType,
		Recalled:    m.Recalled,
		Edited:      m.Edited,
		CreatedAt:   m.CreatedAt,
	}
	if m.QuoteMsgId != 0 {
		item.QuoteMsgId = strconv.FormatInt(m.QuoteMsgId, 10)
	}
	if m.Extra != "" {
		item.Extra = m.Extra
	}
	return item
}
