// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessagesLogic {
	return &SearchMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchMessagesLogic) SearchMessages(req *types.SearchMessagesReq) (resp *types.SearchMessagesResp, err error) {
	// todo: add your logic here and delete this line

	return
}
