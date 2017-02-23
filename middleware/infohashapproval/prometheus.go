package infohashapproval

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Request
	chihayaAnnounceCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_announce_total_count",
		Help: "Amount of announces the middleware recieves",
	})

	chihayaAnnounceWhitelistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_whitelist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in the whitelist",
	})

	chihayaAnnounceBlacklistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_blacklist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in the blacklist",
	})

	chihayaAnnounceNolistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_nolist_announce_total_count",
		Help: "Amount of announces the middleware recieves that are in no list",
	})

	// Storage
	chihayaWhitelistCount = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_whitelist_total_count",
		Help: "Amount of whitlisted infohashes in the middleware",
	})

	chihayaWhitelistFail = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "chihaya_whitelist_fail_total_count",
		Help: "Amount of whitlisted infohashes failed to write to the middleware",
	})
)

func InitPrometheus() {
	// Request
	prometheus.MustRegister(chihayaAnnounceCount)
	prometheus.MustRegister(chihayaAnnounceWhitelistCount)
	prometheus.MustRegister(chihayaAnnounceBlacklistCount)
	prometheus.MustRegister(chihayaAnnounceNolistCount)

	// Storage
	prometheus.MustRegister(chihayaWhitelistCount)
	prometheus.MustRegister(chihayaWhitelistFail)
}
