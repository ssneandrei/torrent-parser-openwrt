package main

import (
 "html"
 "regexp"
 "strconv"
 "strings"
 "time"
)

var (
 rtRowsRE = regexp.MustCompile(`(?is)<tr\b[^>]*class=["'][^"']*tCenter[^"']*["'][^>]*>(.*?)</tr>`)
 rtTitleRE = regexp.MustCompile(`(?is)<a\b[^>]*class=["'][^"']*tLink[^"']*["'][^>]*>.*?</a>`)
 rtDownRE = regexp.MustCompile(`(?is)<a\b[^>]*class=["'][^"']*tr-dl[^"']*["'][^>]*>.*?</a>`)
 rtCellRE = regexp.MustCompile(`(?is)<td\b([^>]*)>(.*?)</td>`)
 htmlTagRE = regexp.MustCompile(`(?is)<[^>]+>`)
 attrRE = regexp.MustCompile(`(?is)\b([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*["']([^"']*)["']`)
 topicRE = regexp.MustCompile(`(?:[?&]t=|data-topic_id=["']?)(\d+)`)
 numberRE = regexp.MustCompile(`\d+`)
)

func parseRuTrackerResults(tracker TrackerConfig, query SearchQuery, page string) []SearchResult {
 rows:= rtRowsRE.FindAllStringSubmatch(page, -1)
 results:= make([]SearchResult, 0, len(rows))

 for _, match:= range rows {
 row:= match[1]
 titleTag:= rtTitleRE.FindString(row)
 if titleTag == "" {
 continue
 }

 details:= absoluteURL(tracker.BaseURL, htmlAttr(titleTag, "href"))
 id:= topicID(titleTag + " " + details)
 if id == "" {
 continue
 }

 result:= SearchResult{
 TrackerID: "rutracker",
 Title: plainHTML(titleTag),
 DetailsURL: details,
 GUID: "rutracker:" + id,
 Categories: append([]int(nil), query.Categories...),
 }

 cells:= rtCellRE.FindAllStringSubmatch(row, -1)
 for _, cell:= range cells {
 attrs, body:= cell[1], cell[2]
 if strings.Contains(htmlAttr("<td "+attrs+">", "class"), "tor-size") {
 if n, err:= strconv.ParseInt(htmlAttr("<td "+attrs+">", "data-ts_text"), 10, 64); err == nil {
 result.Size = n
 }
 if tag:= rtDownRE.FindString(body); tag!= "" {
 result.TorrentURL = absoluteURL(tracker.BaseURL, htmlAttr(tag, "href"))
 }
 }
 }

 if len(cells) >= 8 {
 result.Seeders = firstInt(plainHTML(cells[6][2]))
 result.Leechers = firstInt(plainHTML(cells[7][2]))
 }
 if len(cells) >= 10 {
 if ts, err:= strconv.ParseInt(htmlAttr("<td "+cells[9][1]+">", "data-ts_text"), 10, 64); err == nil && ts > 0 {
 result.PublishDate = time.Unix(ts, 0)
 }
 }
 if len(result.Categories) == 0 {
 result.Categories = []int{8000}
 }

 results = append(results, result)
 if len(results) >= 200 {
 break
 }
 }

 return results
}

func htmlAttr(tag, name string) string {
 for _, m:= range attrRE.FindAllStringSubmatch(tag, -1) {
 if strings.EqualFold(m[1], name) {
 return html.UnescapeString(m[2])
 }
 }
 return ""
}

func plainHTML(s string) string {
 s = htmlTagRE.ReplaceAllString(s, " ")
 return strings.Join(strings.Fields(html.UnescapeString(s)), " ")
}

func topicID(s string) string {
 if m:= topicRE.FindStringSubmatch(s); m!= nil {
 return m[1]
 }
 return ""
}

func firstInt(s string) int {
 v:= numberRE.FindString(s)
 n, _:= strconv.Atoi(v)
 return n
}
