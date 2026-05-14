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

type GetConversationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetConversationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConversationsLogic {
	return &GetConversationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetConversationsLogic) GetConversations(req *types.GetConversationsReq) (resp *types.GetConversationsResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	result, err := l.svcCtx.ConversationRpc.GetConversations(l.ctx, &conversationclient.GetConversationsReq{
		Uid: uid,
	})
	if err != nil {
		l.Errorf("GetConversations RPC failed: %v", err)
		return nil, err
	}

	var peerUids []int64
	for _, item := range result.List {
		if item.Type == "SINGLE" {
			peerUid := item.PeerUid
			if item.Uid == uid {
				peerUid = item.PeerUid
			} else {
				peerUid = item.Uid
			}
			if peerUid > 0 {
				peerUids = append(peerUids, peerUid)
			}
		}
	}

	userMap := make(map[int64]*userclient.UserInfo)
	if len(peerUids) > 0 {
		batchResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userclient.BatchGetUserInfoReq{
			Uids: peerUids,
		})
		if err != nil {
			l.Errorf("BatchGetUserInfo failed: %v", err)
		} else {
			userMap = batchResp.Users
		}
	}

	list := make([]types.ConversationItem, len(result.List))
	for i, item := range result.List {
		name := item.Name
		avatar := item.Avatar
		if item.Type == "SINGLE" {
			peerUid := item.PeerUid
			if item.Uid == uid {
				peerUid = item.PeerUid
			} else {
				peerUid = item.Uid
			}
			if peerUid > 0 {
				if u, ok := userMap[peerUid]; ok {
					name = u.Name
					avatar = u.Avatar
				}
			}
		}

		list[i] = types.ConversationItem{
			ConvID:         strconv.FormatInt(item.ConvId, 10),
			Type:           item.Type,
			Name:           name,
			Avatar:         avatar,
			GroupID:        strconv.FormatInt(item.GroupId, 10),
			LastMsgID:      strconv.FormatInt(item.LastMsgId, 10),
			LastMsgSnippet: item.LastMsgSnippet,
			LastMsgTime:    item.LastMsgTime,
			LastMsgSender:  strconv.FormatInt(item.LastMsgSender, 10),
			UnreadCount:    item.UnreadCount,
			Mute:           item.Mute,
			Pinned:         item.Pinned,
			Uid:            strconv.FormatInt(item.Uid, 10),
			PeerUid:        strconv.FormatInt(item.PeerUid, 10),
		}
	}
	return &types.GetConversationsResp{
		List: list,
	}, nil
}
