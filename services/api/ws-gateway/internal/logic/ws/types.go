package ws

import "encoding/json"

type WSMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
	Seq  int64           `json:"seq,omitempty"`
}

type AuthData struct {
	Token string `json:"token"`
}

type AuthResp struct {
	Success bool   `json:"success"`
	Uid     string `json:"uid,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SendMessageData struct {
	ConvId      string `json:"conv_id"`
	Receiver    string `json:"receiver,omitempty"`
	Type        string `json:"msg_type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	QuoteMsgId  string `json:"quote_msg_id,omitempty"`
	Extra       string `json:"extra,omitempty"`
	TempId      string `json:"temp_id,omitempty"`
}

type NewMessagePush struct {
	MsgId        string `json:"msg_id"`
	ConvId       string `json:"conv_id"`
	Sender       string `json:"sender"`
	SenderName   string `json:"sender_name"`
	SenderAvatar string `json:"sender_avatar"`
	Receiver     string `json:"receiver,omitempty"`
	Type         string `json:"msg_type"`
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	QuoteMsgId   string `json:"quote_msg_id,omitempty"`
	Extra        string `json:"extra,omitempty"`
	CreatedAt    int64  `json:"created_at"`
}

type ReadReceiptData struct {
	ConvId        string `json:"conv_id"`
	LastReadMsgId string `json:"last_read_msg_id"`
	StartMsgId    string `json:"start_msg_id,omitempty"`
	EndMsgId      string `json:"end_msg_id,omitempty"`
}

type ReadReceiptPush struct {
	ConvId        string `json:"conv_id"`
	Uid           string `json:"uid"`
	LastReadMsgId string `json:"last_read_msg_id"`
	StartMsgId    string `json:"start_msg_id,omitempty"`
	EndMsgId      string `json:"end_msg_id,omitempty"`
}

type TypingData struct {
	ConvId   string `json:"conv_id"`
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
}

type TypingPush struct {
	ConvId string `json:"conv_id"`
	Uid    string `json:"uid"`
	Status string `json:"status"`
}

type RecallMessageData struct {
	ConvId string `json:"conv_id"`
	MsgId  string `json:"msg_id"`
}

type RecallMessagePush struct {
	ConvId string `json:"conv_id"`
	MsgId  string `json:"msg_id"`
	Uid    string `json:"uid"`
}

type EditMessageData struct {
	ConvId  string `json:"conv_id"`
	MsgId   string `json:"msg_id"`
	Content string `json:"content"`
}

type EditMessagePush struct {
	ConvId  string `json:"conv_id"`
	MsgId   string `json:"msg_id"`
	Content string `json:"content"`
	Uid     string `json:"uid"`
}

type OnlineStatusPush struct {
	Uid    string `json:"uid"`
	Status string `json:"status"`
}

type ErrorPush struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
}

type SearchMessagesData struct {
	Keyword   string `json:"keyword"`
	ConvId    string `json:"conv_id,omitempty"`
	Sender    string `json:"sender,omitempty"`
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Page      int32  `json:"page"`
	Size      int32  `json:"size"`
}

type SearchMessagesResp struct {
	List  []*MsgItem `json:"list"`
	Total int64      `json:"total"`
}

type MsgItem struct {
	Id          string `json:"id"`
	ConvId      string `json:"conv_id"`
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	QuoteMsgId  string `json:"quote_msg_id,omitempty"`
	Recalled    bool   `json:"recalled"`
	Edited      bool   `json:"edited"`
	Extra       string `json:"extra,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

type SyncOfflineData struct {
	Limit int32 `json:"limit"`
}

type SyncOfflineResp struct {
	List []*MsgItem `json:"list"`
}

type GetMessagesData struct {
	ConvId    string `json:"conv_id"`
	Cursor    string `json:"cursor"`
	Limit     int32  `json:"limit"`
	Direction string `json:"direction"`
}

type GetMessagesResp struct {
	List    []*MsgItem `json:"list"`
	HasMore bool       `json:"has_more"`
}

type GetReadMembersData struct {
	ConvId string `json:"conv_id"`
	MsgId  string `json:"msg_id"`
}

type ReadMemberItem struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	ReadAt int64  `json:"read_at"`
}

type GetReadMembersResp struct {
	List  []*ReadMemberItem `json:"list"`
	Total int64             `json:"total"`
}

type BotIdentifyData struct {
	Token string `json:"token"`
}

type BotIdentifyResp struct {
	Success bool   `json:"success"`
	BotId   string `json:"bot_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BotMessageEvent struct {
	EventId     string `json:"event_id"`
	EventType   string `json:"event_type"`
	ConvId      string `json:"conv_id"`
	MsgId       string `json:"msg_id"`
	Sender      string `json:"sender"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	QuoteMsgId  string `json:"quote_msg_id,omitempty"`
	Extra       string `json:"extra,omitempty"`
	CreatedAt   int64  `json:"created_at"`
}

type BotMentionEvent struct {
	EventId      string `json:"event_id"`
	EventType    string `json:"event_type"`
	ConvId       string `json:"conv_id"`
	MsgId        string `json:"msg_id"`
	Sender       string `json:"sender"`
	SenderName   string `json:"sender_name"`
	SenderAvatar string `json:"sender_avatar"`
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	MentionName  string `json:"mention_name"`
	QuoteMsgId   string `json:"quote_msg_id,omitempty"`
	CreatedAt    int64  `json:"created_at"`
}

type BotSendMessageData struct {
	ConvId       string `json:"conv_id"`
	MsgType      string `json:"msg_type"`
	Content      string `json:"content"`
	ContentType  string `json:"content_type"`
	SenderName   string `json:"sender_name,omitempty"`
	SenderAvatar string `json:"sender_avatar,omitempty"`
	QuoteMsgId   string `json:"quote_msg_id,omitempty"`
	Extra        string `json:"extra,omitempty"`
}

type BotSendMessageResp struct {
	MsgId     string `json:"msg_id"`
	ConvId    string `json:"conv_id"`
	CreatedAt int64  `json:"created_at"`
	Error     string `json:"error,omitempty"`
}

type BotRecallMessageData struct {
	ConvId string `json:"conv_id"`
	MsgId  string `json:"msg_id"`
}

type BotRecallMessageResp struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

type BotGetMessageData struct {
	ConvId string `json:"conv_id"`
	MsgId  string `json:"msg_id"`
}

type BotGetMessageResp struct {
	Msg   *MsgItem `json:"msg,omitempty"`
	Error string   `json:"error,omitempty"`
}

type BotGetConversationData struct {
	ConvId string `json:"conv_id"`
}

type BotGetConversationResp struct {
	ConvId  string           `json:"conv_id"`
	Type    string           `json:"type"`
	Name    string           `json:"name"`
	Members []*BotConvMember `json:"members,omitempty"`
	Error   string           `json:"error,omitempty"`
}

type BotConvMember struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Role   string `json:"role"`
}

type BotGetUserData struct {
	Uid string `json:"uid"`
}

type BotGetUserResp struct {
	Uid    string `json:"uid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Error  string `json:"error,omitempty"`
}

type BotUploadFileData struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	FileSize    int64  `json:"file_size"`
}

type BotUploadFileResp struct {
	FileId  string `json:"file_id,omitempty"`
	FileUrl string `json:"file_url,omitempty"`
	Error   string `json:"error,omitempty"`
}

type BotHeartbeatData struct {
	Seq int64 `json:"seq,omitempty"`
}

type BotHello struct {
	HeartbeatInterval int64 `json:"heartbeat_interval"`
}

type BotReady struct {
	SessionId string `json:"session_id"`
	BotId     string `json:"bot_id"`
}

func NewWSMessage(msgType string, data interface{}) (*WSMessage, error) {
	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	return &WSMessage{
		Type: msgType,
		Data: raw,
	}, nil
}

func MustNewWSMessage(msgType string, data interface{}) *WSMessage {
	msg, err := NewWSMessage(msgType, data)
	if err != nil {
		msg = &WSMessage{Type: msgType}
	}
	return msg
}
