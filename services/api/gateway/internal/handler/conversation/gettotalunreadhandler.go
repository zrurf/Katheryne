// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package conversation

import (
	"net/http"

	"gateway/internal/logic/conversation"
	"gateway/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func GetTotalUnreadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := conversation.NewGetTotalUnreadLogic(r.Context(), svcCtx)
		resp, err := l.GetTotalUnread()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
