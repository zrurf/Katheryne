package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"bot/botclient"
	"ws-gateway/internal/logic/ws"
	"ws-gateway/internal/svc"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ClientWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}

		uid, err := svcCtx.Redis.Get(r.Context(), "access_token:"+token).Int64()
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("ws upgrade error: %v", err)
			return
		}

		tokenPrefix := token
		if len(tokenPrefix) > 8 {
			tokenPrefix = tokenPrefix[:8]
		}
		logx.Infof("ws upgrade success: uid=%d, token_prefix=%s", uid, tokenPrefix)

		client := ws.NewClient(uid, svcCtx.Hub, conn)
		svcCtx.Hub.Register(client)

		authResp := &ws.AuthResp{
			Success: true,
			Uid:     strconv.FormatInt(uid, 10),
		}
		authMsg := ws.MustNewWSMessage("auth_resp", authResp)
		authData, _ := json.Marshal(authMsg)
		client.SendMessage(authData)

		go client.WritePump()
		go client.ReadPump()
	}
}

func BotWSHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logx.Errorf("bot ws upgrade error: %v", err)
			return
		}

		hello := &ws.BotHello{
			HeartbeatInterval: svcCtx.Config.WSHeartbeatInterval,
		}
		helloMsg := ws.MustNewWSMessage("hello", hello)
		helloData, _ := json.Marshal(helloMsg)
		if err := conn.WriteMessage(websocket.TextMessage, helloData); err != nil {
			logx.Errorf("bot ws hello error: %v", err)
			conn.Close()
			return
		}

		_, identifyData, err := conn.ReadMessage()
		if err != nil {
			logx.Errorf("bot ws read identify error: %v", err)
			conn.Close()
			return
		}

		var msg ws.WSMessage
		if err := json.Unmarshal(identifyData, &msg); err != nil || msg.Type != "identify" {
			errResp := ws.MustNewWSMessage("error", &ws.ErrorPush{Code: 401, Message: "identify required"})
			errData, _ := json.Marshal(errResp)
			conn.WriteMessage(websocket.TextMessage, errData)
			conn.Close()
			return
		}

		var identify ws.BotIdentifyData
		if err := json.Unmarshal(msg.Data, &identify); err != nil {
			errResp := ws.MustNewWSMessage("error", &ws.ErrorPush{Code: 400, Message: "invalid identify data"})
			errData, _ := json.Marshal(errResp)
			conn.WriteMessage(websocket.TextMessage, errData)
			conn.Close()
			return
		}

		botId, err := resolveBotID(svcCtx, identify.Token)
		if err != nil {
			errResp := ws.MustNewWSMessage("error", &ws.ErrorPush{Code: 401, Message: "invalid bot token"})
			errData, _ := json.Marshal(errResp)
			conn.WriteMessage(websocket.TextMessage, errData)
			conn.Close()
			return
		}

		botClient := ws.NewBotClient(botId, svcCtx.Hub, conn)
		svcCtx.Hub.RegisterBot(botClient)

		ready := &ws.BotReady{
			SessionId: "",
			BotId:     strconv.FormatInt(botId, 10),
		}
		readyMsg := ws.MustNewWSMessage("ready", ready)
		readyData, _ := json.Marshal(readyMsg)
		botClient.SendMessage(readyData)

		go botClient.WritePump()
		go botClient.ReadPump()
	}
}

type botTokenInfo struct {
	BotID     int64  `json:"bot_id"`
	ClientID  string `json:"client_id"`
	Scope     string `json:"scope"`
	ExpiresAt int64  `json:"expires_at"`
}

func resolveBotID(svcCtx *svc.ServiceContext, token string) (int64, error) {
	// Try bot_access_token: prefix (OAuth2 client_credentials grant)
	jsonData, err := svcCtx.Redis.Get(context.Background(), "bot_access_token:"+token).Result()
	if err == nil {
		var info botTokenInfo
		if err := json.Unmarshal([]byte(jsonData), &info); err == nil {
			return info.BotID, nil
		}
	}

	// Try bot_token: prefix (legacy)
	botID, err := svcCtx.Redis.Get(context.Background(), "bot_token:"+token).Int64()
	if err == nil {
		return botID, nil
	}

	// Try oauth2:token: prefix (gateway-issued token for client_credentials)
	clientID, err := svcCtx.Redis.HGet(context.Background(), "oauth2:token:"+token, "client_id").Result()
	if err == nil && clientID != "" {
		resp, rpcErr := svcCtx.BotRpc.ResolveBotCredential(context.Background(), &botclient.ResolveBotCredentialReq{ClientId: clientID})
		if rpcErr == nil && resp != nil {
			return resp.BotId, nil
		}
	}

	return 0, err
}
