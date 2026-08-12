package main

import (
 "bytes"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "regexp"
 "strconv"
 "strings"
)

const maxTorrentBody = 16 << 20

var (
 rtDownloadRE = regexp.MustCompile(`href=["'](dl\.php\?t=\d+)["']`)
 magnetHashRE = regexp.MustCompile(`(?i)magnet:\?xt=urn:btih:([a-f0-9]{40})`)
)

func (a *App) handleDownload(w http.ResponseWriter, r *http.Request) {
 if r.Method!= http.MethodGet {
 http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
 return
 }

 if!a.authorized(r) {
 http.Error(w, "unauthorized", http.StatusUnauthorized)
 return
 }

 parts:= strings.Split(strings.TrimPrefix(r.URL.Path, "/download/"), "/")
 if len(parts)!= 2 || parts[0] == "" || parts[1] == "" {
 http.Error(w, "bad download path", http.StatusBadRequest)
 return
 }

 trackerID, id:= parts[0], parts[1]
 if _, err:= strconv.Atoi(id); err!= nil {
 http.Error(w, "bad id", http.StatusBadRequest)
 return
 }

 tracker, ok:= a.cfg.Trackers[trackerID]
 if!ok ||!tracker.Enabled {
 http.Error(w, "tracker unavailable", http.StatusNotFound)
 return
 }

 var data []byte
 var filename string
 var magnet string
 var err error

 switch trackerID {
 case "rutracker":
 data, filename, magnet, err = a.downloadRuTracker(tracker, id)
 case "kinozal":
 data, filename, err = a.downloadKinozal(tracker, id)
 case "nnmclub":
 magnet, err = a.downloadNNMClub(tracker, id)
 default:
 http.Error(w, "tracker unavailable", http.StatusNotFound)
 return
 }

 if err!= nil {
 http.Error(w, err.Error(), http.StatusBadGateway)
 return
 }

 if magnet!= "" {
 http.Redirect(w, r, magnet, http.StatusFound)
 return
 }

 if!looksTorrent(data) {
 http.Error(w, "tracker returned non-torrent payload", http.StatusBadGateway)
 return
 }

 w.Header().Set("Content-Type", "application/x-bittorrent")
 w.Header().Set(
 "Content-Disposition",
 fmt.Sprintf(`attachment; filename="%s"`, filename),
 )
 w.Header().Set("Content-Length", strconv.Itoa(len(data)))
 _, _ = w.Write(data)
}

func (a *App) downloadRuTracker(
 tracker TrackerConfig,
 id string,
) ([]byte, string, string, error) {
 client, err:= newSessionClient(a.client)
 if err!= nil {
 return nil, "", "", err
 }

 if tracker.Username!= "" && tracker.Password!= "" {
 if err:= loginRuTracker(client, tracker); err!= nil {
 return nil, "", "", err
 }
 }

 baseURL:= strings.TrimRight(tracker.BaseURL, "/")
 page, err:= a.downloadGET(
 client,
 baseURL+"/forum/viewtopic.php?t="+id,
 4<<20,
 )
 if err!= nil {
 return nil, "", "", err
 }

 if tracker.Username!= "" && tracker.Password!= "" {
 if match:= rtDownloadRE.FindSubmatch(page); match!= nil {
 torrent, err:= a.downloadGET(
 client,
 baseURL+"/forum/"+string(match[1]),
 maxTorrentBody,
 )
 if err == nil && looksTorrent(torrent) {
 return torrent, "rutracker-" + id + ".torrent", "", nil
 }
 }
 }

 if magnet:= magnetFromPage(page); magnet!= "" {
 return nil, "", magnet, nil
 }

 return nil, "", "", errors.New("no torrent or magnet on topic page")
}

func (a *App) downloadKinozal(
 tracker TrackerConfig,
 id string,
) ([]byte, string, error) {
 client, err:= newSessionClient(a.client)
 if err!= nil {
 return nil, "", err
 }

 if err:= loginKinozal(client, tracker); err!= nil {
 return nil, "", err
 }

 baseURL:= strings.TrimRight(tracker.BaseURL, "/")
 parsed, err:= url.Parse(baseURL)
 if err!= nil {
 return nil, "", err
 }

 targets:= []string{
 "https://dl." + parsed.Host + "/download.php?id=" + id,
 baseURL + "/download.php?id=" + id,
 }

 var lastErr error

 for _, target:= range targets {
 data, err:= a.downloadGET(client, target, maxTorrentBody)
 if err == nil && looksTorrent(data) {
 return data, "kinozal-" + id + ".torrent", nil
 }
    if err!= nil {
 lastErr = err
 } else {
 lastErr = errors.New("non-torrent response")
 }
 }

 if lastErr == nil {
 lastErr = errors.New("torrent download failed")
 }

 return nil, "", lastErr
}

func (a *App) downloadNNMClub(
 tracker TrackerConfig,
 id string,
) (string, error) {
 page, err:= a.downloadGET(
 a.client,
 strings.TrimRight(tracker.BaseURL, "/")+"/forum/viewtopic.php?t="+id,
 4<<20,
 )
 if err!= nil {
 return "", err
 }

 magnet:= magnetFromPage(page)
 if magnet == "" {
 return "", errors.New("topic page has no magnet")
 }

 return magnet, nil
}

func (a *App) downloadGET(
 client *http.Client,
 target string,
 limit int64,
) ([]byte, error) {
 req, err:= http.NewRequest(http.MethodGet, target, nil)
 if err!= nil {
 return nil, err
 }

 userAgent:= a.cfg.UserAgent
 if userAgent == "" {
 userAgent = defaultUA
 }
 req.Header.Set("User-Agent", userAgent)

 resp, err:= client.Do(req)
 if err!= nil {
 return nil, err
 }
 defer resp.Body.Close()

 data, err:= io.ReadAll(io.LimitReader(resp.Body, limit+1))
 if err!= nil {
 return nil, err
 }

 if int64(len(data)) > limit {
 return nil, fmt.Errorf("response is larger than %d bytes", limit)
 }

 if resp.StatusCode < 200 || resp.StatusCode >= 400 {
 return nil, fmt.Errorf("GET %s: HTTP %d", target, resp.StatusCode)
 }

 return data, nil
}

func magnetFromPage(page []byte) string {
 match:= magnetHashRE.FindSubmatch(page)
 if match == nil {
 return ""
 }

 return "magnet:?xt=urn:btih:" + strings.ToLower(string(match[1]))
}

func looksTorrent(data []byte) bool {
 if len(data) <= 20 || data[0]!= 'd' {
 return false
 }

 limit:= len(data)
 if limit > 4096 {
 limit = 4096
 }

 return bytes.Contains(data[:limit], []byte("4:info"))
}
