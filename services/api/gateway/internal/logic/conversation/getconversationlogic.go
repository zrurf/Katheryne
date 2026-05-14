package conversation

import (
	"context"
	"strconv"

	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"
	"user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetConversationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConversationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationLogic {
	return &GetConversationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationLogic) GetConversation(req *types.GetConversationReq) (resp *types.GetConversationResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.ConversationRpc.GetConversation(l.ctx, &conversationclient.GetConversationReq{
		ConvId: convId,
		Uid:    uid,
	})
	if err != nil {
		l.Errorf("GetConversation RPC failed: %v", err)
		return nil, err
	}

	name := result.Name
	avatar := result.Avatar
	peerUidStr := ""

	if result.Type == "SINGLE" {
		peerUid := result.PeerUid
		if result.Uid == uid {
			peerUid = result.PeerUid
		} else {
			peerUid = result.Uid
		}
		peerUidStr = strconv.FormatInt(peerUid, 10)

		if peerUid > 0 {
			userResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userclient.BatchGetUserInfoReq{
				Uids: []int64{peerUid},
			})
			if err == nil && userResp != nil {
				if u, ok := userResp.Users[peerUid]; ok {
					name = u.Name
					avatar = u.Avatar
				}
			}
		}
	}

	return &types.GetConversationResp{
		ConvID:  strconv.FormatInt(result.ConvId, 10),
		Type:    result.Type,
		Name:    name,
		Avatar:  avatar,
		GroupID: strconv.FormatInt(result.GroupId, 10),
		Mute:    result.Mute,
		Pinned:  result.Pinned,
		PeerUid: peerUidStr,
	}, nil
}
