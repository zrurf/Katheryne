// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oauth2

import (
	"net/http"

	"bot/internal/logic/oauth2"
	"bot/internal/svc"
	"bot/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func BotTokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BotTokenReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := oauth2.NewBotTokenLogic(r.Context(), svcCtx)
		resp, err := l.BotToken(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
