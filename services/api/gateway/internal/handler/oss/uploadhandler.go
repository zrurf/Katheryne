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
		// Enforce file size limit from config
		maxSize := svcCtx.Config.MaxFileSize
		if maxSize <= 0 {
			maxSize = 104857600 // 100 MB default
		}
		if r.ContentLength > maxSize {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			xhttp.JsonBaseResponseCtx(r.Context(), w, nil)
			return
		}
		if err := r.ParseMultipartForm(maxSize); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		l := oss.NewUploadLogic(r.Context(), svcCtx)
		l.SetFile(file)
		l.SetFileName(header.Filename)
		l.SetContentType(header.Header.Get("Content-Type"))
		l.SetHost(r.Host)

		resp, err := l.Upload()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}
