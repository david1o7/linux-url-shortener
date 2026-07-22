package validator

import (
	"Linux-url-shortener/internal/logger"
	"context"
	"net/http"
	"net/url"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type URLValidator struct {
	client HTTPClient
	resolver DNSResolver
	timeout time.Duration
}

func NewURLValidator(client HTTPClient, resolver DNSResolver, timeout int) *URLValidator {
    if client == nil {
        client = &http.Client{
            Timeout: time.Duration(timeout) * time.Second,
        }
    }

    if resolver == nil {
        resolver = &RealResolver{}
    }

    return &URLValidator{
        client:   client,
        resolver: resolver,
        timeout:  time.Duration(timeout) * time.Second,
    }
}

func (v *URLValidator) Validate(rawURL string) bool {

	u, err := url.ParseRequestURI(rawURL)

	if err != nil {
		return false
	}

	if !isAllowedScheme(u.Scheme) {
		return false
	}

	host := u.Hostname()

	if host == "" {
		return false
	}

	if v.isPrivateOrLoopback(host) {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodHead,
		rawURL,
		nil,
	)

	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "LinuxURLShortener/1.0")

	resp, err := v.client.Do(req)
	if err != nil {
		logger.Log.Info(
			"HTTP ERROR",
			"HTTP ERROR", err.Error(),
		)
		return false
	}

	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 400
}