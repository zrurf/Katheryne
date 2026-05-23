package bot_interact

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	summarizeTimeout  = 120 * time.Second
	summarizeMsgLimit = 60
)

type BotSummarizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotSummarizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotSummarizeLogic {
	return &BotSummarizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotSummarizeLogic) BotSummarize(req *types.BotSummarizeReq) (resp *types.SummarizeTicket, err error) {
	ticket := uuid.New().String()

	convID, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		l.Errorf("invalid conv_id: %s", req.ConvID)
		return nil, err
	}

	// Determine cursor: use since_msg_id from request, or from Redis cache
	cursor := req.SinceMsgId
	if cursor == 0 {
		// Check Redis for last summarized msg_id for this conversation
		cached, redisErr := l.svcCtx.Redis.Get(l.ctx, "summarize:cursor:"+req.ConvID).Result()
		if redisErr == nil {
			cursor, _ = strconv.ParseInt(cached, 10, 64)
		}
	}

	// Fetch only new messages since cursor (or latest N if no cursor)
	var msgs *messageclient.GetMessagesResp
	if cursor > 0 {
		msgs, err = l.svcCtx.MessageRpc.GetMessages(l.ctx, &messageclient.GetMessagesReq{
			ConvId:    convID,
			Cursor:    cursor,
			Direction: "after",
			Limit:     summarizeMsgLimit,
		})
	} else {
		msgs, err = l.svcCtx.MessageRpc.GetMessages(l.ctx, &messageclient.GetMessagesReq{
			ConvId: convID,
			Limit:  summarizeMsgLimit,
		})
	}
	if err != nil {
		l.Errorf("GetMessages RPC failed: %v", err)
		return nil, err
	}

	if len(msgs.List) == 0 {
		l.Errorf("no messages to summarize for conv %s", req.ConvID)
		return nil, errNoMessages
	}

	// Build chat messages for ai-bot
	chatMsgs := make([]map[string]string, 0, len(msgs.List))
	for _, m := range msgs.List {
		role := "user"
		chatMsgs = append(chatMsgs, map[string]string{
			"role":    role,
			"content": m.Content,
		})
	}

	// Store pending status in Redis with timeout
	err = l.svcCtx.Redis.Set(l.ctx, "summarize:"+ticket, `{"status":"processing"}`, summarizeTimeout).Err()
	if err != nil {
		l.Errorf("Redis set failed: %v", err)
		return nil, err
	}

	// Async: call ai-bot and update result
	go l.doSummarize(ticket, req.ConvID, chatMsgs, msgs.List)

	return &types.SummarizeTicket{Ticket: ticket}, nil
}

func (l *BotSummarizeLogic) doSummarize(ticket string, convID string, chatMsgs []map[string]string, msgs []*messageclient.MsgItem) {
	defer func() {
		if r := recover(); r != nil {
			l.Errorf("panicked in doSummarize: %v", r)
			l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
				`{"status":"error","error":"internal error"}`, summarizeTimeout)
		}
	}()

	url := l.svcCtx.Config.AiBotUrl + "/bot/summarize"
	body, err := json.Marshal(map[string]interface{}{
		"messages": chatMsgs,
	})
	if err != nil {
		l.Errorf("marshal summarize request failed: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), summarizeTimeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		l.Errorf("create summarize request failed: %v", err)
		l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
			`{"status":"error","error":"`+err.Error()+`"}`, summarizeTimeout)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: summarizeTimeout}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		l.Errorf("ai-bot summarize request failed: %v", err)
		l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
			`{"status":"error","error":"`+err.Error()+`"}`, summarizeTimeout)
		return
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		l.Errorf("read summarize response failed: %v", err)
		l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
			`{"status":"error","error":"`+err.Error()+`"}`, summarizeTimeout)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		l.Errorf("ai-bot summarize returned %d: %s", httpResp.StatusCode, string(respBody))
		l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
			`{"status":"error","error":"ai-bot error: `+string(respBody)+`"}`, summarizeTimeout)
		return
	}

	var wrapper struct {
		Code int                     `json:"code"`
		Msg  string                  `json:"msg"`
		Data *types.BotSummarizeResp `json:"data"`
	}
	if err := json.Unmarshal(respBody, &wrapper); err != nil {
		l.Errorf("unmarshal summarize response failed: %v", err)
		l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket,
			`{"status":"error","error":"`+err.Error()+`"}`, summarizeTimeout)
		return
	}

	resultJSON, _ := json.Marshal(map[string]interface{}{
		"status":       "completed",
		"summary":      wrapper.Data.Summary,
		"key_points":   wrapper.Data.KeyPoints,
		"action_items": wrapper.Data.ActionItems,
	})
	l.svcCtx.Redis.Set(context.Background(), "summarize:"+ticket, string(resultJSON), summarizeTimeout)

	// Cache the last summarized msg_id for this conversation
	lastMsgId := msgs[len(msgs)-1].Id
	l.svcCtx.Redis.Set(context.Background(), "summarize:cursor:"+convID, strconv.FormatInt(lastMsgId, 10), 7*24*time.Hour)
}

var errNoMessages = &noMessagesError{}

type noMessagesError struct{}

func (e *noMessagesError) Error() string { return "no messages to summarize" }
