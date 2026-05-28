package message

import (
	"context"
	"strconv"

	"bot/botclient"
	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"
	"user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessagesLogic {
	return &GetMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessagesLogic) GetMessages(req *types.GetMessagesReq) (resp *types.GetMessagesResp, err error) {
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		l.Errorf("GetMessages invalid conv_id: %s, err=%v", req.ConvID, err)
		return nil, err
	}
	var cursor int64
	if req.Cursor != "" {
		cursor, err = strconv.ParseInt(req.Cursor, 10, 64)
		if err != nil {
			l.Errorf("GetMessages invalid cursor: %s, err=%v", req.Cursor, err)
			return nil, err
		}
	}

	l.Infof("GetMessages calling MessageRpc.GetMessages: convId=%d, cursor=%d, limit=%d, direction=%s", convId, cursor, req.Limit, req.Direction)
	result, err := l.svcCtx.MessageRpc.GetMessages(l.ctx, &messageclient.GetMessagesReq{
		ConvId:    convId,
		Cursor:    cursor,
		Limit:     int32(req.Limit),
		Direction: req.Direction,
	})
	if err != nil {
		l.Errorf("GetMessages RPC failed: convId=%d, err=%v", convId, err)
		return nil, err
	}
	l.Infof("GetMessages RPC success: convId=%d, count=%d, hasMore=%v", convId, len(result.List), result.HasMore)

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

	botMap := make(map[int64]*botclient.InstalledBotItem)
	botResp, botErr := l.svcCtx.BotRpc.GetConvBots(l.ctx, &botclient.GetConvBotsReq{ConvId: convId})
	if botErr == nil {
		for _, bot := range botResp.List {
			botMap[bot.BotId] = bot
		}
	}

	list := make([]types.MessageItem, len(result.List))
	for i, item := range result.List {
		senderName := ""
		senderAvatar := ""
		if u, ok := userMap[item.Sender]; ok {
			senderName = u.Name
			senderAvatar = u.Avatar
		} else if bot, ok := botMap[item.Sender]; ok {
			senderName = bot.Name
			senderAvatar = bot.Avatar
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
	return &types.GetMessagesResp{
		List:    list,
		HasMore: result.HasMore,
	}, nil
}
