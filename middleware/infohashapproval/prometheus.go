package infohashapproval

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Request
	//		Announce
	chihayaAnnounceCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_announce_total_count",
		Help: "Amount of announces the middleware recieves",
	})

	chihayaAnnounceWhitelistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_whitelist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in the whitelist",
	})

	chihayaAnnounceBlacklistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_blacklist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in the blacklist",
	})

	chihayaAnnounceNolistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_nolist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in no list",
	})

	chihayaAnnounceResponseTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Name: "chihaya_middleware_announce_time_summary_ns",
		Help: "Announce request times",
	})

	//		Scrape
	chihayaScrapCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_scrape_count",
		Help: "Number of scrape requests",
	})

	// Storage
	chihayaWhitelistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_whitelist_total_count",
		Help: "Amount of whitlisted infohashes in the middleware",
	})

	chihayaWhitelistFail = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_middleware_whitelist_fail_total_count",
		Help: "Amount of whitlisted infohashes failed to write to the middleware",
	})
)

func InitPrometheus() {
	// Request
	prometheus.MustRegister(chihayaAnnounceCount)
	prometheus.MustRegister(chihayaAnnounceWhitelistCount)
	prometheus.MustRegister(chihayaAnnounceBlacklistCount)
	prometheus.MustRegister(chihayaAnnounceNolistCount)
	prometheus.MustRegister(chihayaScrapCount)
	prometheus.MustRegister(chihayaAnnounceResponseTime)

	// Storage
	prometheus.MustRegister(chihayaWhitelistCount)
	prometheus.MustRegister(chihayaWhitelistFail)
}
