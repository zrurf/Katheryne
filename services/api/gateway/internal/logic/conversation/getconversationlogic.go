// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationLogic {
	return &GetConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationLogic) GetConversation(req *types.GetConversationReq) (resp *types.GetConversationResp, err error) {
	// todo: add your logic here and delete this line

	return
}
