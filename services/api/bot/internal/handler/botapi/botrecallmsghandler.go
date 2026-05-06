// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"net/http"

	"bot/internal/logic/botapi"
	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func BotRecallMsgHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BotRecallMsgReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := botapi.NewBotRecallMsgLogic(r.Context(), svcCtx)
		resp, err := l.BotRecallMsg(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
