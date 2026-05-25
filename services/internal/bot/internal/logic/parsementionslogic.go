package logic

import (
	"context"
	"regexp"
	"strconv"

	"bot/bot"
	"bot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

var mentionRegex = regexp.MustCompile(`@\[(bot|user):(\d+):([^\]]+)\]`)

type ParseMentionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewParseMentionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ParseMentionsLogic {
	return &ParseMentionsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ParseMentionsLogic) ParseMentions(in *bot.ParseMentionsReq) (*bot.ParseMentionsResp, error) {
	matches := mentionRegex.FindAllStringSubmatchIndex(in.Content, -1)
	var mentions []*bot.MentionItem

	for _, m := range matches {
		if len(m) < 8 {
			continue
		}
		mentionType := in.Content[m[2]:m[3]]
		idStr := in.Content[m[4]:m[5]]
		displayName := in.Content[m[6]:m[7]]

		targetID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}

		mentions = append(mentions, &bot.MentionItem{
			Type:        mentionType,
			TargetId:    targetID,
			DisplayName: displayName,
			StartPos:    int32(m[0]),
			EndPos:      int32(m[1]),
		})
	}

	return &bot.ParseMentionsResp{Mentions: mentions}, nil
}