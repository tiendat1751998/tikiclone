package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	LivestreamsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_live_created_total", Help: "Total livestreams created",
	})
	LivestreamsStarted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_live_started_total", Help: "Total livestreams started",
	})
	LivestreamsEnded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_live_ended_total", Help: "Total livestreams ended",
	})
	ChatMessagesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_live_chat_total", Help: "Total chat messages sent",
	})
	ReactionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_live_reactions_total", Help: "Total reactions by type",
	}, []string{"type"})
	GiftsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_live_gifts_total", Help: "Total gifts by type",
	}, []string{"gift_type"})
	WSConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_live_ws_connections_active", Help: "Active WebSocket connections",
	})
	WSConnectionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_live_ws_connections_total", Help: "Total WebSocket connections",
	})
	FanoutLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "tiki_live_fanout_latency_ms",
		Help:    "Fanout latency in milliseconds",
		Buckets: []float64{1, 5, 10, 25, 50, 100, 250, 500},
	})
	ModerationActionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_live_moderation_actions_total", Help: "Moderation actions by type",
	}, []string{"action"})
	ViewerCountGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_live_viewer_count", Help: "Viewer count by room",
	}, []string{"room_id"})

	ConnectionsActive = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tiki_live_ws_connections_active_by_room", Help: "Active WebSocket connections by room",
	}, []string{"room_id"})

	MessagesBroadcast = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tiki_live_messages_broadcast_total", Help: "Messages broadcast by type",
	}, []string{"type"})
)
