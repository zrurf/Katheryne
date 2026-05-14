// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package developer

import (
	"net/http"

	"bot/internal/logic/developer"
	"bot/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func ListMyBotsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := developer.NewListMyBotsLogic(r.Context(), svcCtx)
		resp, err := l.ListMyBots()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
