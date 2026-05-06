// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package botapi

import (
	"net/http"

	"bot/internal/logic/botapi"
	"bot/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func BotUploadFileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := botapi.NewBotUploadFileLogic(r.Context(), svcCtx)
		resp, err := l.BotUploadFile()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
