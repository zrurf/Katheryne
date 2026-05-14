package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OnlineUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "online_users",
		Help:      "Current number of online users",
	})

	WsConnectionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "connections_total",
		Help:      "Total number of WebSocket connections",
	}, []string{"type"})

	WsConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "connections_active",
		Help:      "Current number of active WebSocket connections",
	})

	WsMessagesReceivedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "messages_received_total",
		Help:      "Total number of messages received via WebSocket",
	}, []string{"type", "msg_type"})

	WsMessagesSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "messages_sent_total",
		Help:      "Total number of messages pushed via WebSocket",
	}, []string{"type", "msg_type"})

	WsBotMessagesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "bot_messages_total",
		Help:      "Total number of bot messages processed",
	}, []string{"action"})

	WsBotMessageLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "bot_message_latency_seconds",
		Help:      "Bot message processing latency in seconds",
		Buckets:   prometheus.DefBuckets,
	}, []string{"action"})

	WsErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "errors_total",
		Help:      "Total number of WebSocket errors",
	}, []string{"type", "code"})

	WsHeartbeatFailures = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "heartbeat_failures_total",
		Help:      "Total number of heartbeat failures",
	})

	WsBroadcastTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "broadcast_total",
		Help:      "Total number of broadcast messages",
	})

	WsBotConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "aim",
		Subsystem: "ws",
		Name:      "bot_connections_active",
		Help:      "Current number of active bot WebSocket connections",
	})
)