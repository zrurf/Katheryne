// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"context"

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
	// todo: add your logic here and delete this line

	return
}
