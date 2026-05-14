package social

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"social/socialclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TransferOwnerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTransferOwnerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TransferOwnerLogic {
	return &TransferOwnerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TransferOwnerLogic) TransferOwner(req *types.TransferOwnerReq) (resp *types.TransferOwnerResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	groupId, err := strconv.ParseInt(req.GroupID, 10, 64)
	if err != nil {
		return nil, err
	}
	newOwner, err := strconv.ParseInt(req.NewOwner, 10, 64)
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.SocialRpc.TransferOwner(l.ctx, &socialclient.TransferOwnerReq{
		GroupId:  groupId,
		OldOwner: uid,
		NewOwner: newOwner,
	})
	if err != nil {
		l.Errorf("TransferOwner RPC failed: %v", err)
		return nil, err
	}
	return &types.TransferOwnerResp{Result: true}, nil
}
