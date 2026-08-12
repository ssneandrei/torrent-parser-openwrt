package main

import (
 "encoding/xml"
 "fmt"
 "log"
 "net/http"
 "net/url"
 "sort"
 "strconv"
 "strings"
 "time"
)

type torznabRSS struct {
 XMLName xml.Name `xml:"rss"`
 Version string `xml:"version,attr"`
 XMLNSTorznab string `xml:"xmlns:torznab,attr"`
 Channel torznabChannel `xml:"channel"`
}

type torznabChannel struct {
 Title string `xml:"title"`
 Description string `xml:"description"`
 Link string `xml:"link"`
 Items []torznabItem `xml:"item"`
}

type torznabItem struct {
 Title string `xml:"title"`
 GUID torznabGUID `xml:"guid"`
 Link string `xml:"link"`
 Comments string `xml:"comments,omitempty"`
 PubDate string `xml:"pubDate,omitempty"`
 Size int64 `xml:"size,omitempty"`
 Categories []int `xml:"category,omitempty"`
 Enclosure *torznabEnclosure `xml:"enclosure,omitempty"`
 Attrs []torznabAttr `xml:"torznab:attr"`
}

type torznabGUID struct {
 IsPermaLink string `xml:"isPermaLink,attr"`
 Value string `xml:",chardata"`
}

type torznabEnclosure struct {
 URL string `xml:"url,attr"`
 Length int64 `xml:"length,attr"`
 Type string `xml:"type,attr"`
}

type torznabAttr struct {
 Name string `xml:"name,attr"`
 Value string `xml:"value,attr"`
}

func (a *App) handleTorznab(w http.ResponseWriter, r *http.Request) {
 if r.Method!= http.MethodGet {
 http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
 return
 }

 if!a.authorized(r) {
 http.Error(w, "unauthorized", http.StatusUnauthorized)
 return
 }

 action:= strings.ToLower(strings.TrimSpace(r.URL.Query().Get("t")))
 if action == "" {
 action = "search"
 }

 switch action {
 case "caps":
 a.writeCaps(w)

 case "search", "tvsearch", "movie", "music", "book":
 query:= parseSearchQuery(r)
 results, errs:= a.searchAll(query)

 for _, err:= range errs {
 log.Printf("tracker search: %v", err)
 }

 a.writeTorznabResults(w, r, query, results)

 default:
 http.Error(w, "unsupported Torznab action", http.StatusBadRequest)
 }
}

func parseSearchQuery(r *http.Request) SearchQuery {
 values:= r.URL.Query()

 limit:= 100
 if n, err:= strconv.Atoi(values.Get("limit")); err == nil && n > 0 {
 limit = n
 }
 if limit > 200 {
 limit = 200
 }

 offset:= 0
 if n, err:= strconv.Atoi(values.Get("offset")); err == nil && n >= 0 {
 offset = n
 }

 categories:= make([]int, 0)

 for _, raw:= range strings.Split(values.Get("cat"), ",") {
 raw = strings.TrimSpace(raw)
 if raw == "" {
 continue
 }

 if n, err:= strconv.Atoi(raw); err == nil && n > 0 {
 categories = append(categories, n)
 }
 }

 return SearchQuery{
 Query: strings.TrimSpace(values.Get("q")),
 Categories: categories,
 Limit: limit,
 Offset: offset,
 }
}

func (a *App) writeCaps(w http.ResponseWriter) {
 w.Header().Set("Content-Type", "application/xml; charset=utf-8")

 fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<caps>
 <server version="%s" title="Torrent Parser OpenWrt" />
 <limits max="200" default="100" />
 <searching>
 <search available="yes" supportedParams="q" />
 <tv-search available="yes" supportedParams="q,season,ep" />
 <movie-search available="yes" supportedParams="q,imdbid" />
 </searching>
 <categories>
 <category id="2000" name="Movies" />
 <category id="5000" name="TV" />
 <category id="8000" name="Other" />
 </categories>
</caps>`, version)
}

func (a *App) writeTorznabResults(
 w http.ResponseWriter,
 r *http.Request,
 query SearchQuery,
 results []SearchResult,
) {
 sort.SliceStable(results, func(i, j int) bool {
 return results[i].PublishDate.After(results[j].PublishDate)
 })

 start:= query.Offset
 if start > len(results) {
 start = len(results)
 }

 end:= start + query.Limit
 if end > len(results) {
 end = len(results)
 }

 selected:= results[start:end]
 items:= make([]torznabItem, 0, len(selected))

 for _, result:= range selected {
 items = append(items, a.resultToTorznabItem(r, result))
 }
   feed:= torznabRSS{
 Version: "2.0",
 XMLNSTorznab: "http://torznab.com/schemas/2015/feed",
 Channel: torznabChannel{
 Title: "Torrent Parser OpenWrt",
 Description: "RuTracker, Kinozal and NNMClub Torznab feed",
 Link: requestBaseURL(r) + "/api",
 Items: items,
 },
 }

 data, err:= xml.MarshalIndent(feed, "", " ")
 if err!= nil {
 http.Error(w, "xml encode error", http.StatusInternalServerError)
 return
 }

 w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
 _, _ = w.Write([]byte(xml.Header))
 _, _ = w.Write(data)
}

func (a *App) resultToTorznabItem(
 r *http.Request,
 result SearchResult,
) torznabItem {
 guid:= strings.TrimSpace(result.GUID)
 if guid == "" {
 guid = result.TrackerID + ":" + result.DetailsURL
 }

 link:= strings.TrimSpace(result.DetailsURL)
 prefix:= result.TrackerID + ":"

 if strings.HasPrefix(guid, prefix) {
 id:= strings.TrimPrefix(guid, prefix)
 if id!= "" {
 link = requestBaseURL(r) +
 "/download/" + url.PathEscape(result.TrackerID) +
 "/" + url.PathEscape(id)

 if a.cfg.APIKey!= "" {
 link += "?[REDACTED] + url.QueryEscape(a.cfg.APIKey)
 }
 }
 }

 categories:= result.Categories
 if len(categories) == 0 {
 categories = []int{8000}
 }

 attrs:= []torznabAttr{
 {Name: "tracker", Value: result.TrackerID},
 }

 if result.Seeders >= 0 {
 attrs = append(attrs, torznabAttr{
 Name: "seeders",
 Value: strconv.Itoa(result.Seeders),
 })
 }

 if result.Seeders >= 0 || result.Leechers >= 0 {
 seeders:= result.Seeders
 leechers:= result.Leechers
 if seeders < 0 {
 seeders = 0
 }
 if leechers < 0 {
 leechers = 0
 }

 attrs = append(attrs, torznabAttr{
 Name: "peers",
 Value: strconv.Itoa(seeders + leechers),
 })
 }

 if result.Size > 0 {
 attrs = append(attrs, torznabAttr{
 Name: "size",
 Value: strconv.FormatInt(result.Size, 10),
 })
 }

 item:= torznabItem{
 Title: result.Title,
 GUID: torznabGUID{
 IsPermaLink: "false",
 Value: guid,
 },
 Link: link,
 Comments: result.DetailsURL,
 Size: result.Size,
 Categories: categories,
 Attrs: attrs,
 }

 if!result.PublishDate.IsZero() {
 item.PubDate = result.PublishDate.UTC().Format(time.RFC1123Z)
 }

 if link!= "" && result.Size > 0 {
 item.Enclosure = &torznabEnclosure{
 URL: link,
 Length: result.Size,
 Type: "application/x-bittorrent",
 }
 }

 return item
}

func requestBaseURL(r *http.Request) string {
 scheme:= "http"
 if r.TLS!= nil {
 scheme = "https"
 }

 host:= r.Host
 return scheme + "://" + host
}
