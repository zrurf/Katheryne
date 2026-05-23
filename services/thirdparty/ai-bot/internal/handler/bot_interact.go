package handler

import (
	"ai-bot/internal/svc"
	"encoding/json"
	"net/http"

	"ai-bot/internal/types"

	xhttp "github.com/zeromicro/x/http"
)

type BotInteract struct {
	svcCtx *svc.ServiceContext
}

func NewBotInteractHandler(svcCtx *svc.ServiceContext) *BotInteract {
	return &BotInteract{svcCtx: svcCtx}
}

func (h *BotInteract) SummarizeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []types.ChatMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		result, err := h.svcCtx.MsgHandler.SummarizeContext(req.Messages)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		xhttp.JsonBaseResponseCtx(r.Context(), w, result)
	}
}

func (h *BotInteract) TranslateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Text       string `json:"text"`
			SourceLang string `json:"source_lang"`
			TargetLang string `json:"target_lang"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		result, err := h.svcCtx.MsgHandler.TranslateText(req.Text, req.SourceLang, req.TargetLang)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		xhttp.JsonBaseResponseCtx(r.Context(), w, result)
	}
}

func (h *BotInteract) SuggestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []types.ChatMessage `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		result, err := h.svcCtx.MsgHandler.SuggestRepliesContext(req.Messages)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		xhttp.JsonBaseResponseCtx(r.Context(), w, result)
	}
}

func (h *BotInteract) ModerateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		result, err := h.svcCtx.MsgHandler.ModerateText(req.Text)
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, err)
			return
		}

		xhttp.JsonBaseResponseCtx(r.Context(), w, result)
	}
}