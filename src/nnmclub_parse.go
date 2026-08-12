package main

import (
 "regexp"
 "strconv"
 "strings"
 "time"
)

var (
 nnmRowRE = regexp.MustCompile(`(?is)<tr\b[^>]*>(.*?)</tr>`)
 nnmTitleRE = regexp.MustCompile(`(?is)<a\b[^>]*href=["'](?:[^"']*/)?viewtopic\.php\?t=(\d+)[^"']*["'][^>]*>\s*(?:<b\b[^>]*>)?(.*?)(?:</b>)?\s*</a>`)
 nnmDownloadRE = regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*download\.php\?id=(\d+)[^"']*)["'][^>]*>`)
 nnmSeedRE = regexp.MustCompile(`(?is)<td\b[^>]*class=["'][^"']*seedmed[^"']*["'][^>]*>.*?<b\b[^>]*>\s*(\d+)`)
 nnmLeechRE = regexp.MustCompile(`(?is)<td\b[^>]*class=["'][^"']*leechmed[^"']*["'][^>]*>.*?<b\b[^>]*>\s*(\d+)`)
 nnmUNumberRE = regexp.MustCompile(`(?is)<u\b[^>]*>\s*(\d+)\s*</u>`)
)

func parseNNMClubResults(
 tracker TrackerConfig,
 query SearchQuery,
 page string,
) []SearchResult {
 rows:= nnmRowRE.FindAllStringSubmatch(page, -1)
 out:= make([]SearchResult, 0)

 for _, match:= range rows {
 row:= match[1]
 title:= nnmTitleRE.FindStringSubmatch(row)
 if title == nil {
 continue
 }

 id:= title[1]

 result:= SearchResult{
 TrackerID: "nnmclub",
 Title: plainHTML(title[2]),
 DetailsURL: strings.TrimRight(tracker.BaseURL, "/") +
 "/forum/viewtopic.php?t=" + id,
 GUID: "nnmclub:" + id,
 Seeders: -1,
 Leechers: -1,
 Categories: append([]int(nil), query.Categories...),
 }

 if download:= nnmDownloadRE.FindStringSubmatch(row); download!= nil {
 result.TorrentURL = absoluteURL(tracker.BaseURL, download[1])
 }

 if seed:= nnmSeedRE.FindStringSubmatch(row); seed!= nil {
 result.Seeders, _ = strconv.Atoi(seed[1])
 }

 if leech:= nnmLeechRE.FindStringSubmatch(row); leech!= nil {
 result.Leechers, _ = strconv.Atoi(leech[1])
 }

 cells:= rtCellRE.FindAllStringSubmatch(row, -1)

 if len(cells) >= 6 {
 if size:= nnmUNumberRE.FindStringSubmatch(cells[5][2]); size!= nil {
 result.Size, _ = strconv.ParseInt(size[1], 10, 64)
 } else {
 result.Size = parseKinozalSize(plainHTML(cells[5][2]))
 }
 }

 if len(cells) > 0 {
 last:= cells[len(cells)-1][2]

 if stamp:= nnmUNumberRE.FindStringSubmatch(last); stamp!= nil {
 if ts, err:= strconv.ParseInt(stamp[1], 10, 64); err == nil && ts > 0 {
 result.PublishDate = time.Unix(ts, 0)
 }
 }
 }

 if len(result.Categories) == 0 {
 result.Categories = []int{8000}
 }

 out = append(out, result)
 if len(out) >= 200 {
 break
 }
 }

 return out
}
