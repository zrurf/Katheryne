package message

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageReq) (resp *types.SendMessageResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	var receiver int64
	if req.Receiver != "" {
		receiver, err = strconv.ParseInt(req.Receiver, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	var quoteMsgId int64
	if req.QuoteMsgID != "" {
		quoteMsgId, err = strconv.ParseInt(req.QuoteMsgID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	result, err := l.svcCtx.MessageRpc.SendMessage(l.ctx, &messageclient.SendMessageReq{
		ConvId:      convId,
		Sender:      uid,
		Receiver:    receiver,
		Type:        req.Type,
		Content:     req.Content,
		ContentType: req.ContentType,
		QuoteMsgId:  quoteMsgId,
		Extra:       req.Extra,
	})
	if err != nil {
		l.Errorf("SendMessage RPC failed: %v", err)
		return nil, err
	}

	// Forward the message to ws-gateway so it can:
	// 1. Broadcast the message to online clients in the conversation
	// 2. Route @[bot:xxx:yyy] mentions to connected bots
	l.pushToWsGateway(uid, req, result)

	return &types.SendMessageResp{
		MsgID:     strconv.FormatInt(result.MsgId, 10),
		ConvID:    strconv.FormatInt(result.ConvId, 10),
		CreatedAt: result.CreatedAt,
	}, nil
}

// pushToWsGateway sends the message to ws-gateway's internal push endpoint
// so that WebSocket-connected clients receive the message and @mentions are
// routed to bots.
// This runs in a goroutine so it doesn't block the REST response.
func (l *SendMessageLogic) pushToWsGateway(sender int64, req *types.SendMessageReq, result *messageclient.SendMessageResp) {
	go func() {
		pushReq := map[string]interface{}{
			"conv_id":      req.ConvID,
			"sender":       strconv.FormatInt(sender, 10),
			"type":         req.Type,
			"content":      req.Content,
			"content_type": req.ContentType,
			"msg_id":       strconv.FormatInt(result.MsgId, 10),
			"created_at":   result.CreatedAt,
		}
		if req.Receiver != "" {
			pushReq["receiver"] = req.Receiver
		}
		if req.QuoteMsgID != "" {
			pushReq["quote_msg_id"] = req.QuoteMsgID
		}
		if req.Extra != "" {
			pushReq["extra"] = req.Extra
		}

		body, _ := json.Marshal(pushReq)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
			l.svcCtx.Config.WsGatewayUrl+"/api/v1/internal/push_message",
			bytes.NewReader(body))
		if err != nil {
			l.Errorf("pushToWsGateway create request error: %v", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			l.Errorf("pushToWsGateway error: convId=%s, msgId=%d, err=%v",
				req.ConvID, result.MsgId, err)
			return
		}
		resp.Body.Close()
	}()
}
