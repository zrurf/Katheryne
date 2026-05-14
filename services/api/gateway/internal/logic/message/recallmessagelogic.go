package message

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecallMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecallMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecallMessageLogic {
	return &RecallMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecallMessageLogic) RecallMessage(req *types.RecallMessageReq) (resp *types.RecallMessageResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	msgId, err := strconv.ParseInt(req.MsgID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.MessageRpc.RecallMessage(l.ctx, &messageclient.RecallMessageReq{
		ConvId:   convId,
		MsgId:    msgId,
		Operator: uid,
	})
	if err != nil {
		l.Errorf("RecallMessage RPC failed: %v", err)
		return nil, err
	}
	return &types.RecallMessageResp{}, nil
}
