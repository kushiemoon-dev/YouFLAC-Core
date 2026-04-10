package core

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoWithProxyFallback(t *testing.T) {
	tests := []struct {
		name          string
		directStatus  int
		proxyAvail    bool
		wantCallsDir  int32
		wantCallsProx int32
		wantStatus    int
		wantErr       bool
	}{
		{"200 direct, no fallback", 200, true, 1, 0, 200, false},
		{"403 direct, proxy recovers", 403, true, 1, 1, 200, false},
		{"429 direct, proxy recovers", 429, true, 1, 1, 200, false},
		{"451 direct, proxy recovers", 451, true, 1, 1, 200, false},
		{"500 direct, no retry", 500, true, 1, 0, 500, false},
		{"403 direct, no proxy configured", 403, false, 1, 0, 403, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var directCalls, proxyCalls int32

			direct := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&directCalls, 1)
				w.WriteHeader(tt.directStatus)
			}))
			defer direct.Close()

			proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				atomic.AddInt32(&proxyCalls, 1)
				w.WriteHeader(200)
			}))
			defer proxy.Close()

			directClient, _ := NewHTTPClient(5*time.Second, "")
			var proxyClient *http.Client
			if tt.proxyAvail {
				proxyClient, _ = NewHTTPClient(5*time.Second, "")
			}

			target := direct.URL
			proxyTarget := proxy.URL

			req, _ := http.NewRequest("GET", target, nil)
			resp, err := DoWithProxyFallback(directClient, proxyClient, req, proxyTarget)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp != nil {
				defer resp.Body.Close()
				if resp.StatusCode != tt.wantStatus {
					t.Errorf("status=%d want=%d", resp.StatusCode, tt.wantStatus)
				}
			}
			if got := atomic.LoadInt32(&directCalls); got != tt.wantCallsDir {
				t.Errorf("direct calls=%d want=%d", got, tt.wantCallsDir)
			}
			if got := atomic.LoadInt32(&proxyCalls); got != tt.wantCallsProx {
				t.Errorf("proxy calls=%d want=%d", got, tt.wantCallsProx)
			}
		})
	}
}

func TestShouldRetryWithProxy(t *testing.T) {
	cases := []struct {
		code int
		want bool
	}{
		{200, false},
		{403, true},
		{429, true},
		{451, true},
		{500, false},
		{404, false},
	}
	for _, c := range cases {
		if got := shouldRetryWithProxy(c.code); got != c.want {
			t.Errorf("code=%d got=%v want=%v", c.code, got, c.want)
		}
	}
}
