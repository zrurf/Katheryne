// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"net/http"

	"gateway/internal/logic/social"
	"gateway/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func GetBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := social.NewGetBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.GetBlacklist()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
