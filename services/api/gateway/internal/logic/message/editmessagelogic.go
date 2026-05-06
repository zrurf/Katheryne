// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package message

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

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
	// todo: add your logic here and delete this line

	return
}
