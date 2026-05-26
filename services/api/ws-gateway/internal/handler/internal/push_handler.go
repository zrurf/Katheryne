package internal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"ws-gateway/internal/logic/ws"
	"ws-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	xhttp "github.com/zeromicro/x/http"
)

// PushMessageReq is the request body for the internal message push endpoint.
// It mirrors ws.BroadcastMsg fields (string-formatted IDs for safe transport).
type PushMessageReq struct {
	ConvId      string `json:"conv_id"`
	Sender      string `json:"sender"`
	Receiver    string `json:"receiver,optional"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	QuoteMsgId  string `json:"quote_msg_id,optional"`
	Extra       string `json:"extra,optional"`
	MsgId       string `json:"msg_id"`
	CreatedAt   int64  `json:"created_at"`
}

// PushMessageHandler creates an HTTP handler that pushes a message to the Hub
// for broadcast. Used by the gateway API to relay REST-sent messages through
// the WebSocket broadcast pipeline, which includes @mention routing to bots.
func PushMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PushMessageReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logx.Errorf("push_message decode error: %v", err)
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
		sender, _ := strconv.ParseInt(req.Sender, 10, 64)
		receiver, _ := strconv.ParseInt(req.Receiver, 10, 64)
		msgId, _ := strconv.ParseInt(req.MsgId, 10, 64)
		quoteMsgId, _ := strconv.ParseInt(req.QuoteMsgId, 10, 64)

		broadcast := &ws.BroadcastMsg{
			ConvId:      convId,
			Sender:      sender,
			Receiver:    receiver,
			MsgType:     req.Type,
			Content:     req.Content,
			ContentType: req.ContentType,
			QuoteMsgId:  quoteMsgId,
			Extra:       req.Extra,
			MsgId:       msgId,
			CreatedAt:   req.CreatedAt,
		}

		svcCtx.Hub.Broadcast(broadcast)

		xhttp.JsonBaseResponseCtx(r.Context(), w, map[string]string{"status": "ok"})
	}
}