package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetConvMuteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetConvMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetConvMuteLogic {
	return &SetConvMuteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetConvMuteLogic) SetConvMute(req *types.SetConvMuteReq) (resp *types.SetConvMuteResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ConversationRpc.SetConvMute(l.ctx, &conversationclient.SetConvMuteReq{
		ConvId: convId,
		Uid:    uid,
		Mute:   req.Mute,
	})
	if err != nil {
		l.Errorf("SetConvMute RPC failed: %v", err)
		return nil, err
	}
	return &types.SetConvMuteResp{}, nil
}
