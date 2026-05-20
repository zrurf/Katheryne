package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveAuthorizeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewApproveAuthorizeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveAuthorizeLogic {
	return &ApproveAuthorizeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ApproveAuthorizeLogic) ApproveAuthorize(req *types.ApproveAuthorizeReq) (resp *types.ApproveAuthorizeResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.ApproveAuthorize(l.ctx, &botclient.ApproveAuthorizeReq{
		ClientId:    req.ClientId,
		RedirectUri: req.RedirectUri,
		Scope:       req.Scope,
		Uid:         uid,
		ConvId:      convId,
		State:       req.State,
	})
	if err != nil {
		return nil, err
	}
	return &types.ApproveAuthorizeResp{
		RedirectUrl: result.RedirectUrl,
	}, nil
}
