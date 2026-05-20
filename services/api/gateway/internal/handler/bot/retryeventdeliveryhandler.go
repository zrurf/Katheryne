// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package bot

import (
	"net/http"

	"gateway/internal/logic/bot"
	"gateway/internal/svc"
	"gateway/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func RetryEventDeliveryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RetryEventDeliveryReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := bot.NewRetryEventDeliveryLogic(r.Context(), svcCtx)
		resp, err := l.RetryEventDelivery(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
