// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package auth

import (
	"net/http"

	"gateway/internal/logic/auth"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func RegisterFinalizeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RegisterFinalizeRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := auth.NewRegisterFinalizeLogic(r.Context(), svcCtx)
		resp, err := l.RegisterFinalize(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
