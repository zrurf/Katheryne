package message

import (
	"context"
	"strconv"

	"bot/botclient"
	"conversation/conversationclient"
	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"
	"user/userclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncOfflineMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncOfflineMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncOfflineMessagesLogic {
	return &SyncOfflineMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncOfflineMessagesLogic) SyncOfflineMessages(req *types.SyncOfflineMessagesReq) (resp *types.SyncOfflineMessagesResp, err error) {
	uid := l.ctx.Value("uid").(int64)
	var lastSyncMsgId int64
	if req.LastSyncMsgID != "" {
		lastSyncMsgId, err = strconv.ParseInt(req.LastSyncMsgID, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	var convIds []int64
	convResp, convErr := l.svcCtx.ConversationRpc.GetConversations(l.ctx, &conversationclient.GetConversationsReq{Uid: uid})
	if convErr != nil {
		l.Errorf("GetConversations for sync offline failed: %v", convErr)
	} else {
		for _, conv := range convResp.List {
			convIds = append(convIds, conv.ConvId)
		}
	}

	result, err := l.svcCtx.MessageRpc.SyncOfflineMsgs(l.ctx, &messageclient.SyncOfflineMsgsReq{
		Uid:           uid,
		LastSyncMsgId: lastSyncMsgId,
		DeviceId:      req.DeviceID,
		Limit:         int32(req.Limit),
		ConvIds:       convIds,
	})
	if err != nil {
		l.Errorf("SyncOfflineMsgs RPC failed: %v", err)
		return nil, err
	}

	senderUids := make(map[int64]bool)
	for _, item := range result.Messages {
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
	for _, cid := range convIds {
		botResp, botErr := l.svcCtx.BotRpc.GetConvBots(l.ctx, &botclient.GetConvBotsReq{ConvId: cid})
		if botErr == nil {
			for _, bot := range botResp.List {
				if _, exists := botMap[bot.BotId]; !exists {
					botMap[bot.BotId] = bot
				}
			}
		}
	}

	list := make([]types.MessageItem, len(result.Messages))
	for i, item := range result.Messages {
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
	return &types.SyncOfflineMessagesResp{
		List:             list,
		HasMore:          result.HasMore,
		NewLastSyncMsgID: strconv.FormatInt(result.NewLastSyncMsgId, 10),
	}, nil
}
