package installation

import (
	"context"

	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConvBotsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConvBotsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConvBotsLogic {
	return &GetConvBotsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConvBotsLogic) GetConvBots(req *types.GetConvBotsReq) (resp *types.GetConvBotsResp, err error) {
	uid := l.ctx.Value("uid").(int64)

	convInfo, err := l.svcCtx.InstallationDao.GetConversation(l.ctx, req.ConvID)
	if err != nil {
		return &types.GetConvBotsResp{List: []types.InstalledBotItem{}}, nil
	}

	if convInfo.ConvType == "GROUP" && convInfo.GroupID > 0 {
		if !l.svcCtx.InstallationDao.IsGroupMember(l.ctx, convInfo.GroupID, uid) {
			return &types.GetConvBotsResp{List: []types.InstalledBotItem{}}, nil
		}
	}

	list, err := l.svcCtx.InstallationDao.ListConvBots(l.ctx, req.ConvID)
	if err != nil {
		return &types.GetConvBotsResp{List: []types.InstalledBotItem{}}, nil
	}

	return &types.GetConvBotsResp{List: list}, nil
}
