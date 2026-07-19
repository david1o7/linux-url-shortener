package metrics

import "github.com/prometheus/client_golang/prometheus"

var UrlsShortened = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name : "urls_shortened_total",
		Help: "Total number of url shortened",
	},
)

var Redirects = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name : "redirect_request_total",
		Help: "Total redirect requests",
	},
)

var CacheHits = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help:"Total cache hits",
	},
)
var CacheMisses = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:"cache_misses_total",
		Help:"Total Cache misses", 
	},
)
var RateLimited = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "rate_limited_request_total",
		Help:"Total rate limited requests",
	},
)

func Init() {
	prometheus.MustRegister(
		UrlsShortened,
		Redirects,
		CacheHits,
		CacheMisses,
		RateLimited,
	)
}
