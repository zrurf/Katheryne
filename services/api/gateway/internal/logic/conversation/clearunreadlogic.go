package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ClearUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearUnreadLogic {
	return &ClearUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearUnreadLogic) ClearUnread(req *types.ClearUnreadReq) (resp *types.ClearUnreadResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ConversationRpc.ClearUnread(l.ctx, &conversationclient.ClearUnreadReq{
		ConvId: convId,
		Uid:    uid,
	})
	if err != nil {
		l.Errorf("ClearUnread RPC failed: %v", err)
		return nil, err
	}
	return &types.ClearUnreadResp{}, nil
}
