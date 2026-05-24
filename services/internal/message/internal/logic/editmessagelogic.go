package logic

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type EditMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEditMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EditMessageLogic {
	return &EditMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// editHistoryEntry 编辑历史条目
type editHistoryEntry struct {
	OldContent string `json:"old_content"`
	EditedAt   int64  `json:"edited_at"`
}

func (l *EditMessageLogic) EditMessage(in *message.EditMessageReq) (*message.EditMessageResp, error) {
	m, err := l.svcCtx.MessageDao.GetMessageById(l.ctx, in.MsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	if m.Sender != in.Editor {
		return nil, errors.New("只能编辑自己发送的消息")
	}

	if m.Recalled {
		return nil, errors.New("已撤回的消息不能编辑")
	}

	// 构建编辑历史
	var editHistory []editHistoryEntry
	// 解析已有的 extra 字段中的编辑历史
	if m.Extra.Valid && m.Extra.String != "" {
		var existing struct {
			EditHistory []editHistoryEntry `json:"edit_history"`
		}
		if jsonErr := json.Unmarshal([]byte(m.Extra.String), &existing); jsonErr == nil && existing.EditHistory != nil {
			editHistory = existing.EditHistory
		}
	}

	// 追加本次编辑记录
	editHistory = append(editHistory, editHistoryEntry{
		OldContent: m.Content,
		EditedAt:   time.Now().UnixMilli(),
	})

	// 构建新的 extra JSON
	extraJSON, err := json.Marshal(map[string]interface{}{
		"edit_history": editHistory,
	})
	if err != nil {
		l.Logger.Errorf("marshal edit history failed: %v", err)
		extraJSON = []byte("{}")
	}

	err = l.svcCtx.MessageDao.EditMessage(l.ctx, in.MsgId, in.Content, string(extraJSON))
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	err = l.svcCtx.RedisDao.DelLastMsgCache(l.ctx, in.ConvId)
	if err != nil {
		l.Logger.Error("del last msg cache error:", err)
	}

	return &message.EditMessageResp{}, nil
}
