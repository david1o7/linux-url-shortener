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
var InvalidUrls = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:"Invalid_urls_total",
		Help: "Total invalid url requests",
	},
)


var RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path"},
	)

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_requests_duration_seconds",
		Help: "How long does each requests take",

		Buckets: prometheus.DefBuckets,
	},

	[]string{
		"method",
		"path",
	},
)
func Init() {
	prometheus.MustRegister(
		UrlsShortened,
		Redirects,
		CacheHits,
		CacheMisses,
		RateLimited,
		InvalidUrls,
		RequestDuration,
		RequestCounter,
	)
}
