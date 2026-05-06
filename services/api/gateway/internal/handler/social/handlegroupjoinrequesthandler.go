// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package social

import (
	"net/http"

	"gateway/internal/logic/social"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func HandleGroupJoinRequestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HandleGroupJoinReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := social.NewHandleGroupJoinRequestLogic(r.Context(), svcCtx)
		resp, err := l.HandleGroupJoinRequest(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
