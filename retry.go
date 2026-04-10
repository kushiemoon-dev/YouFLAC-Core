package core

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

// shouldRetryWithProxy returns true when a response code indicates that
// a proxy retry might succeed (geo-block, rate-limit, unavailable-for-legal).
func shouldRetryWithProxy(status int) bool {
	switch status {
	case http.StatusForbidden, http.StatusTooManyRequests, http.StatusUnavailableForLegalReasons:
		return true
	}
	return false
}

// DoWithProxyFallback executes req with directClient. If the response code
// matches shouldRetryWithProxy AND proxyClient is non-nil, it retries the
// same request (rewritten to proxyTarget if non-empty) through proxyClient.
// The original direct response body is closed before the retry.
func DoWithProxyFallback(directClient, proxyClient *http.Client, req *http.Request, proxyTarget string) (*http.Response, error) {
	if directClient == nil {
		return nil, fmt.Errorf("directClient is nil")
	}

	resp, err := directClient.Do(req)
	if err != nil {
		if proxyClient == nil {
			return nil, err
		}
		slog.Warn("direct request failed, retrying via proxy", "err", err)
		return doProxied(proxyClient, req, proxyTarget)
	}

	if !shouldRetryWithProxy(resp.StatusCode) || proxyClient == nil {
		return resp, nil
	}

	slog.Warn("direct request blocked, retrying via proxy",
		"status", resp.StatusCode, "url", req.URL.String())
	resp.Body.Close()

	return doProxied(proxyClient, req, proxyTarget)
}

func doProxied(proxyClient *http.Client, req *http.Request, proxyTarget string) (*http.Response, error) {
	retryReq := req.Clone(req.Context())
	if proxyTarget != "" {
		if u, err := url.Parse(proxyTarget); err == nil {
			retryReq.URL.Scheme = u.Scheme
			retryReq.URL.Host = u.Host
			retryReq.Host = u.Host
		}
	}
	return proxyClient.Do(retryReq)
}
