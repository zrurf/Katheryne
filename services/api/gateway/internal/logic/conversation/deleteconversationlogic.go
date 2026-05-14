package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteConversationLogic {
	return &DeleteConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteConversationLogic) DeleteConversation(req *types.DeleteConversationReq) (resp *types.DeleteConversationResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.ConversationRpc.DeleteConversation(l.ctx, &conversationclient.DeleteConversationReq{
		ConvId: convId,
		Uid:    uid,
	})
	if err != nil {
		l.Errorf("DeleteConversation RPC failed: %v", err)
		return nil, err
	}
	return &types.DeleteConversationResp{}, nil
}
