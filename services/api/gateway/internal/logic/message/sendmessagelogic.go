package message

import (
	"context"
	"strconv"

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

	return &types.SendMessageResp{
		MsgID:     strconv.FormatInt(result.MsgId, 10),
		ConvID:    strconv.FormatInt(result.ConvId, 10),
		CreatedAt: result.CreatedAt,
	}, nil
}
