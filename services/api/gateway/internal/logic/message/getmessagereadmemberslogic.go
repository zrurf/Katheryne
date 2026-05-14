package message

import (
	"context"
	"strconv"

	"gateway/internal/svc"
	"gateway/internal/types"
	"message/messageclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageReadMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMessageReadMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageReadMembersLogic {
	return &GetMessageReadMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessageReadMembersLogic) GetMessageReadMembers(req *types.GetMessageReadMembersReq) (resp *types.GetMessageReadMembersResp, err error) {
	convId, err := strconv.ParseInt(req.ConvID, 10, 64)
	if err != nil {
		return nil, err
	}
	msgId, err := strconv.ParseInt(req.MsgID, 10, 64)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MessageRpc.GetReadMembers(l.ctx, &messageclient.GetReadMembersReq{
		ConvId: convId,
		MsgId:  msgId,
	})
	if err != nil {
		l.Errorf("GetReadMembers RPC failed: %v", err)
		return nil, err
	}

	list := make([]types.ReadMemberItem, len(result.List))
	for i, item := range result.List {
		list[i] = types.ReadMemberItem{
			UID:    strconv.FormatInt(item.Uid, 10),
			Name:   item.Name,
			Avatar: item.Avatar,
			ReadAt: item.ReadAt,
		}
	}
	return &types.GetMessageReadMembersResp{
		List:  list,
		Total: result.Total,
	}, nil
}
