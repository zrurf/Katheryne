package oss

import (
	"net/http"

	"gateway/internal/logic/oss"
	"gateway/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func GetConfigHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oss.NewGetConfigLogic(r.Context(), svcCtx)
		resp, err := l.GetConfig()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
