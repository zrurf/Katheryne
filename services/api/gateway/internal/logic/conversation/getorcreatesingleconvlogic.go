package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrCreateSingleConvLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOrCreateSingleConvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrCreateSingleConvLogic {
	return &GetOrCreateSingleConvLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOrCreateSingleConvLogic) GetOrCreateSingleConv(req *types.GetOrCreateSingleConvReq) (resp *types.GetOrCreateSingleConvResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	peerUid, err := strconv.ParseInt(req.PeerUID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.ConversationRpc.GetOrCreateSingleConv(l.ctx, &conversationclient.GetOrCreateSingleConvReq{
		Uid:     uid,
		PeerUid: peerUid,
	})
	if err != nil {
		l.Errorf("GetOrCreateSingleConv RPC failed: %v", err)
		return nil, err
	}
	return &types.GetOrCreateSingleConvResp{
		ConvID:  strconv.FormatInt(result.ConvId, 10),
		Created: result.Created,
	}, nil
}
