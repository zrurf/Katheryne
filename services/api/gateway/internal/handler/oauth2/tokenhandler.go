// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package oauth2

import (
	"encoding/json"
	"net/http"

	"gateway/internal/logic/oauth2"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func TokenHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.TokenRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauth2.NewTokenLogic(r.Context(), svcCtx)
		resp, err := l.Token(&req)
		if err != nil {
			writeOAuth2Error(w, http.StatusBadRequest, "invalid_grant", err.Error())
			return
		}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func writeOAuth2Error(w http.ResponseWriter, statusCode int, errorCode, description string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}
