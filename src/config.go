package main

import (
 "strconv"
 "strings"
 "time"

 "os/exec"
)

type Config struct {
 Enabled bool
 Listen string
 APIKey string
 Timeout time.Duration
 UserAgent string
 Proxy string
 Trackers map[string]TrackerConfig
}

type TrackerConfig struct {
 ID string
 Enabled bool
 BaseURL string
 Username string
 Password string
}

func loadConfig() (Config, error) {
 cfg:= Config{
 Enabled: true,
 Listen: "0.0.0.0:9696",
 Timeout: 15 * time.Second,
 UserAgent: defaultUA,
 Trackers: make(map[string]TrackerConfig),
 }

 cfg.Enabled = uciBool("torrent-parser.main.enabled", true)

if v:= uciGet("torrent-parser.main.listen"); v!= "" {
 cfg.Listen = v
}

cfg.[REDACTED] v:= uciGet("torrent-parser.main.timeout"); v!= "" {
 if n, err:= strconv.Atoi(v); err == nil && n > 0 {
 cfg.Timeout = time.Duration(n) * time.Second
 }
}


if v:= uciGet("torrent-parser.main.user_agent"); v!= "" {
 cfg.UserAgent = v
}

cfg.Proxy = uciGet("torrent-parser.main.proxy")
 cfg.Trackers["rutracker"] = loadTracker(
 "rutracker",
 "https://rutracker.org",
 )

 cfg.Trackers["kinozal"] = loadTracker(
 "kinozal",
 "https://kinozal.me",
 )

 cfg.Trackers["nnmclub"] = loadTracker(
 "nnmclub",
 "https://nnmclub.to",
 )

 return cfg, nil
}

func loadTracker(id, defaultURL string) TrackerConfig {
 prefix:= "torrent-parser." + id + "."

 t:= TrackerConfig{
 ID: id,
 Enabled: uciBool(prefix+"enabled", true),
 BaseURL: defaultURL,
 }

 if v:= uciGet(prefix + "base_url"); v!= "" {
 t.BaseURL = strings.TrimRight(v, "/")
 }

 t.Username = uciGet(prefix + "username")
 t.Password = uciGet(prefix + "password")

 return t
}

func uciGet(key string) string {
 out, err:= exec.Command("uci", "-q", "get", key).Output()
 if err!= nil {
 return ""
 }

 return strings.TrimSpace(string(out))
}

func uciBool(key string, fallback bool) bool {
 value:= strings.ToLower(uciGet(key))

 switch value {
 case "1", "true", "yes", "on":
 return true
 case "0", "false", "no", "off":
 return false
 default:
 return fallback
 }
}
