package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvPinLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetConvPinLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvPinLogic {
	return &SetConvPinLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetConvPinLogic) SetConvPin(req *types.SetConvPinReq) (resp *types.SetConvPinResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ConversationRpc.SetConvPin(l.ctx, &conversationclient.SetConvPinReq{
		ConvId: convId,
		Uid:    uid,
		Pinned: req.Pinned,
	})
	if err != nil {
		l.Errorf("SetConvPin RPC failed: %v", err)
		return nil, err
	}
	return &types.SetConvPinResp{}, nil
}
