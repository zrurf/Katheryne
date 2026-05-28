package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/logic/auth"
	"gateway/internal/svc"
	"gateway/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
	xhttp "github.com/zeromicro/x/http"
)

func UpdateProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateProfileReq
		if err := httpx.Parse(r, &req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		// Extract and validate Bearer token from Authorization header.
		// Auth routes don't use AuthMiddleware (login/register need to work without token),
		// so we must manually validate the token here for authenticated endpoints.
		uid, err := extractUIDFromToken(svcCtx, r)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, fmt.Errorf("unauthorized"))
			return
		}

		ctx := context.WithValue(r.Context(), "uid", uid)
		l := auth.NewUpdateProfileLogic(ctx, svcCtx)
		resp, err := l.UpdateProfile(&req)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
		} else {
			xhttp.JsonBaseResponseCtx(r.Context(), w, resp)
		}
	}
}

// extractUIDFromToken validates the Bearer token from the request and returns the uid.
// Uses the same access_token:<token> Redis lookup as AuthMiddleware.
func extractUIDFromToken(svcCtx *svc.ServiceContext, r *http.Request) (int64, error) {
	authHeader := r.Header.Get("Authorization")
	parts := strings.Fields(authHeader)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return 0, fmt.Errorf("missing or malformed Authorization header")
	}
	token := parts[1]

	uid, err := svcCtx.Redis.Get(r.Context(), "access_token:"+token).Int64()
	if err != nil {
		return 0, fmt.Errorf("invalid or expired token")
	}
	if uid <= 0 {
		return 0, fmt.Errorf("invalid uid in token")
	}
	return uid, nil
}
