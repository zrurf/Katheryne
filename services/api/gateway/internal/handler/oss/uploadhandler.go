// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"net/http"

	"gateway/internal/logic/oss"
	"gateway/internal/svc"

	xhttp "github.com/zeromicro/x/http"
)

func UploadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := oss.NewUploadLogic(r.Context(), svcCtx)
		resp, err := l.Upload()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
