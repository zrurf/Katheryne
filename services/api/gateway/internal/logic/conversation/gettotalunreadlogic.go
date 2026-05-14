package conversation

import (
	"context"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTotalUnreadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTotalUnreadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTotalUnreadLogic {
	return &GetTotalUnreadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTotalUnreadLogic) GetTotalUnread() (resp *types.GetTotalUnreadResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.ConversationRpc.GetTotalUnread(l.ctx, &conversationclient.GetTotalUnreadReq{
		Uid: uid,
	})
	if err != nil {
		l.Errorf("GetTotalUnread RPC failed: %v", err)
		return nil, err
	}
	return &types.GetTotalUnreadResp{Count: result.Count}, nil
}
