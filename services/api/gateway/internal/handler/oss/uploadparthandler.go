// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"net/http"

	"gateway/internal/logic/oss"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func UploadPartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UploadPartRequest
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		l := oss.NewUploadPartLogic(r.Context(), svcCtx)
		resp, err := l.UploadPart(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
