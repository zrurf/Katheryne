package message

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"
	"user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessagesLogic {
	return &SearchMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchMessagesLogic) SearchMessages(req *types.SearchMessagesReq) (resp *types.SearchMessagesResp, err error) {
	var convId int64
	if req.ConvID != "" {
		convId, err = strconv.ParseInt(req.ConvID, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	var sender int64
	if req.Sender != "" {
		sender, err = strconv.ParseInt(req.Sender, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	result, err := l.svcCtx.MessageRpc.SearchMessages(l.ctx, &messageclient.SearchMessagesReq{
		ConvId:    convId,
		Keyword:   req.Keyword,
		Sender:    sender,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Page:      int32(req.Page),
		Size:      int32(req.Size),
	})
	if err != nil {
		l.Errorf("SearchMessages RPC failed: %v", err)
		return nil, err
	}

	senderUids := make(map[int64]bool)
	for _, item := range result.List {
		if item.Sender > 0 {
			senderUids[item.Sender] = true
		}
	}
	var uids []int64
	for uid := range senderUids {
		uids = append(uids, uid)
	}

	userMap := make(map[int64]*userclient.UserInfo)
	if len(uids) > 0 {
		batchResp, err := l.svcCtx.UserRpc.BatchGetUserInfo(l.ctx, &userclient.BatchGetUserInfoReq{
			Uids: uids,
		})
		if err != nil {
			l.Errorf("BatchGetUserInfo failed: %v", err)
		} else {
			userMap = batchResp.Users
		}
	}

	list := make([]types.MessageItem, len(result.List))
	for i, item := range result.List {
		senderName := ""
		senderAvatar := ""
		if u, ok := userMap[item.Sender]; ok {
			senderName = u.Name
			senderAvatar = u.Avatar
		}

		list[i] = types.MessageItem{
			ID:           strconv.FormatInt(item.Id, 10),
			ConvID:       strconv.FormatInt(item.ConvId, 10),
			Sender:       strconv.FormatInt(item.Sender, 10),
			SenderName:   senderName,
			SenderAvatar: senderAvatar,
			Type:         item.Type,
			Content:      item.Content,
			ContentType:  item.ContentType,
			Recalled:     item.Recalled,
			Edited:       item.Edited,
			Extra:        item.Extra,
			CreatedAt:    item.CreatedAt,
		}
		if item.QuoteMsgId > 0 {
			list[i].QuoteMsgID = strconv.FormatInt(item.QuoteMsgId, 10)
		}
	}
	return &types.SearchMessagesResp{
		List:  list,
		Total: result.Total,
	}, nil
}
