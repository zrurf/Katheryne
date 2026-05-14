package message

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type EditMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEditMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EditMessageLogic {
	return &EditMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EditMessageLogic) EditMessage(req *types.EditMessageReq) (resp *types.EditMessageResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	msgId, err := strconv.ParseInt(req.MsgID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.MessageRpc.EditMessage(l.ctx, &messageclient.EditMessageReq{
		ConvId:  convId,
		MsgId:   msgId,
		Content: req.Content,
		Editor:  uid,
	})
	if err != nil {
		l.Errorf("EditMessage RPC failed: %v", err)
		return nil, err
	}
	return &types.EditMessageResp{}, nil
}
