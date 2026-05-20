package bot

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BotGetConvMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBotGetConvMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BotGetConvMembersLogic {
	return &BotGetConvMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BotGetConvMembersLogic) BotGetConvMembers(req *types.BotGetConvMembersReq) (resp *types.BotGetConvMembersResp, err error) {
	convId, _ := strconv.ParseInt(req.ConvId, 10, 64)
	result, err := l.svcCtx.BotRpc.BotGetConvMembers(l.ctx, &botclient.BotGetConvMembersReq{
		ConvId: convId,
	})
	if err != nil {
		return nil, err
	}

	members := make([]types.MemberItem, 0, len(result.Members))
	for _, m := range result.Members {
		members = append(members, types.MemberItem{
			Uid:    strconv.FormatInt(m.Uid, 10),
			Name:   m.Name,
			Avatar: m.Avatar,
			Role:   m.Role,
		})
	}
	return &types.BotGetConvMembersResp{Members: members}, nil
}
