package main

import (
 "errors"
 "log"
 "net/http"
 "time"
)

const (
 version = "0.1.0-alpha"
 defaultUA = "TorrentParserOpenWrt/0.1"
)

func main() {
 cfg, err:= loadConfig()
 if err!= nil {
 log.Fatalf("config: %v", err)
 }

 if cfg.Listen == "" {
 cfg.Listen = "0.0.0.0:9696"
 }

 app, err:= newApp(cfg)
 if err!= nil {
 log.Fatalf("startup: %v", err)
 }

 mux:= http.NewServeMux()
 mux.HandleFunc("/health", app.handleHealth)
 mux.HandleFunc("/api/v1/status", app.handleStatus)
 mux.HandleFunc("/api", app.handleTorznab)
 mux.HandleFunc("/download/", app.handleDownload)

 srv:= &http.Server{
 Addr: cfg.Listen,
 Handler: requestLog(mux),
 ReadHeaderTimeout: 10 * time.Second,
 IdleTimeout: 60 * time.Second,
 }

 log.Printf("torrent-parser %s listening on %s", version, cfg.Listen)

 if err:= srv.ListenAndServe(); err!= nil &&!errors.Is(err, http.ErrServerClosed) {
 log.Fatal(err)
 }
}
