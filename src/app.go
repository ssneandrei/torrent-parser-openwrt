package main

import (
 "encoding/json"
 "fmt"
 "log"
 "net/http"
 "net/url"
 "strings"
 "time"
)

type App struct {
 cfg Config
 client *http.Client
 startedAt time.Time
}

func newApp(cfg Config) (*App, error) {
 transport:= http.DefaultTransport.(*http.Transport).Clone()

 if strings.TrimSpace(cfg.Proxy)!= "" {
 proxyURL, err:= url.Parse(cfg.Proxy)
 if err!= nil {
 return nil, fmt.Errorf("invalid proxy URL: %w", err)
 }

 switch strings.ToLower(proxyURL.Scheme) {
 case "http", "https", "socks5", "socks5h":
 transport.Proxy = http.ProxyURL(proxyURL)
 default:
 return nil, fmt.Errorf("unsupported proxy scheme: %s", proxyURL.Scheme)
 }
 }

 client:= &http.Client{
 Transport: transport,
 Timeout: cfg.Timeout,
 }

 return &App{
 cfg: cfg,
 client: client,
 startedAt: time.Now(),
 }, nil
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
 if r.Method!= http.MethodGet {
 http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
 return
 }

 writeJSON(w, http.StatusOK, map[string]any{
 "ok": true,
 "version": version,
 })
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
 if r.Method!= http.MethodGet {
 http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
 return
 }

 if!a.authorized(r) {
 http.Error(w, "unauthorized", http.StatusUnauthorized)
 return
 }

 trackers:= make(map[string]any)

 for id, tracker:= range a.cfg.Trackers {
 trackers[id] = map[string]any{
 "enabled": tracker.Enabled,
 "base_url": tracker.BaseURL,
 "has_username": tracker.Username!= "",
 "has_password": tracker.Password!= "",
 }
 }

 writeJSON(w, http.StatusOK, map[string]any{
 "ok": true,
 "version": version,
 "listen": a.cfg.Listen,
 "proxy": a.cfg.Proxy!= "",
 "uptime_sec": int64(time.Since(a.startedAt).Seconds()),
 "trackers": trackers,
 })
}

func (a *App) authorized(r *http.Request) bool {
 if a.cfg.[REDACTED] "" {
 return true
 }

 if r.URL.Query().Get("apikey") == a.cfg.APIKey {
 return true
 }

 auth:= strings.TrimSpace(r.Header.Get("Authorization"))
 const prefix = "Bearer "

 if strings.HasPrefix(auth, prefix) {
 return strings.TrimSpace(strings.TrimPrefix(auth, prefix)) == a.cfg.APIKey
 }

 return false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
 w.Header().Set("Content-Type", "application/json; charset=utf-8")
 w.WriteHeader(status)

 if err:= json.NewEncoder(w).Encode(value); err!= nil {
 log.Printf("json response: %v", err)
 }
}

func requestLog(next http.Handler) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
 started:= time.Now()
 next.ServeHTTP(w, r)

 log.Printf(
 "%s %s from=%s duration=%s",
 r.Method,
 r.URL.Path,
 r.RemoteAddr,
 time.Since(started).Round(time.Millisecond),
 )
 })
}
