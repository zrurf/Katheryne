package logic

import (
	"context"

	"message/internal/svc"
	"message/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReadMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReadMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReadMembersLogic {
	return &GetReadMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReadMembersLogic) GetReadMembers(in *message.GetReadMembersReq) (*message.GetReadMembersResp, error) {
	// todo: add your logic here and delete this line

	return &message.GetReadMembersResp{}, nil
}
