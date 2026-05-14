package logic

import (
	"context"

	"message/internal/svc"
	"message/message"
	"user/userclient"

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
	intervals, err := l.svcCtx.MessageDao.GetReadMembersByMsgId(l.ctx, in.ConvId, in.MsgId)
	if err != nil {
		l.Logger.Error(err)
		return nil, err
	}

	uidMap := make(map[int64]int64)
	for _, r := range intervals {
		if r.EndMsgId >= in.MsgId {
			if t, ok := uidMap[r.Uid]; !ok || r.CreatedAt.UnixMilli() < t {
				uidMap[r.Uid] = r.CreatedAt.UnixMilli()
			}
		}
	}

	uids := make([]int64, 0, len(uidMap))
	for uid := range uidMap {
		uids = append(uids, uid)
	}

	userMap := make(map[int64]*userclient.UserInfo)
	if len(uids) > 0 {
		batchResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userclient.BatchGetUserInfoReq{Uids: uids})
		if err != nil {
			l.Logger.Errorf("BatchGetUserInfo failed: %v", err)
		} else {
			for _, u := range batchResp.Users {
				userMap[u.Uid] = u
			}
		}
	}

	items := make([]*message.ReadMemberItem, 0, len(uidMap))
	for uid, readAt := range uidMap {
		name := ""
		avatar := ""
		if u, ok := userMap[uid]; ok {
			name = u.Name
			avatar = u.Avatar
		}
		items = append(items, &message.ReadMemberItem{
			Uid:    uid,
			Name:   name,
			Avatar: avatar,
			ReadAt: readAt,
		})
	}

	return &message.GetReadMembersResp{
		List:  items,
		Total: int64(len(items)),
	}, nil
}
