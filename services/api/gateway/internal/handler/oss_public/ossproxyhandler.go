package oss_public

import (
	"net/http"

	"gateway/internal/logic/oss_public"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

func OssProxyHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if key == "" {
			logx.Errorf("OssProxy: missing key in request %s", r.URL.String())
			http.Error(w, "missing key parameter", http.StatusBadRequest)
			return
		}
		logx.Infof("OssProxy: serving key=%s", key)

		req := &types.OssProxyRequest{Key: key}
		l := oss_public.NewOssProxyLogic(r.Context(), svcCtx)
		l.SetWriter(w)
		l.SetRequest(r)
		if err := l.OssProxy(req); err != nil {
			logx.Errorf("OssProxy: logic error: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
