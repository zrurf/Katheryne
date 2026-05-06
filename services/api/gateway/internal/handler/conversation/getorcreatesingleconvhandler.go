// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"net/http"

	"gateway/internal/logic/conversation"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func GetOrCreateSingleConvHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetOrCreateSingleConvReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := conversation.NewGetOrCreateSingleConvLogic(r.Context(), svcCtx)
		resp, err := l.GetOrCreateSingleConv(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
