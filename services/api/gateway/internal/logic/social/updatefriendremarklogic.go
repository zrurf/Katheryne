// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"context"

	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateFriendRemarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFriendRemarkLogic {
	return &UpdateFriendRemarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFriendRemarkLogic) UpdateFriendRemark(req *types.UpdateFriendRemarkReq) (resp *types.UpdateFriendRemarkResp, err error) {
	// todo: add your logic here and delete this line

	return
}
