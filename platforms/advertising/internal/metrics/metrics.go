package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	AuctionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_auctions_total",
		Help: "Total number of auctions run",
	})

	ImpressionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_impressions_total",
		Help: "Total number of ad impressions",
	})

	ClicksTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_clicks_total",
		Help: "Total number of ad clicks",
	})

	ConversionsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_conversions_total",
		Help: "Total number of ad conversions",
	})

	SpendTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_spend_total",
		Help: "Total ad spend",
	})

	WinRate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_ad_win_rate",
		Help: "Auction win rate",
	})

	CTR = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_ad_ctr",
		Help: "Click-through rate",
	})

	CampaignsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "tiki_ad_campaigns_active",
		Help: "Number of active campaigns",
	})

	CampaignsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_campaigns_created_total", Help: "Total campaigns created",
	})

	AdRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_requests_total", Help: "Total ad requests",
	})

	ImpressionsRecorded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_impressions_recorded_total", Help: "Total impressions recorded",
	})

	ClicksRecorded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tiki_ad_clicks_recorded_total", Help: "Total clicks recorded",
	})
)
