// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oss

import (
	"net/http"
	"strconv"

	"gateway/internal/logic/oss"
	"gateway/internal/svc"
	"gateway/internal/types"

	xhttp "github.com/zeromicro/x/http"
)

func UploadPartHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query params (raw binary body, not JSON)
		uploadID := r.URL.Query().Get("upload_id")
		partNumStr := r.URL.Query().Get("part_number")
		partNumber, err := strconv.Atoi(partNumStr)
		if err != nil || uploadID == "" {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		req := &types.UploadPartRequest{
			UploadID:   uploadID,
			PartNumber: partNumber,
		}

		l := oss.NewUploadPartLogic(r.Context(), svcCtx)
		l.SetBody(r.Body)
		resp, err := l.UploadPart(req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
