package main

import (
 "regexp"
 "strconv"
 "strings"
 "time"
)

var (
 kzRowRE = regexp.MustCompile(`(?is)<tr\b[^>]*>(.*?)</tr>`)
 kzLinkRE = regexp.MustCompile(`(?is)<a\b[^>]*href=["']([^"']*details\.php\?[^"']*)["'][^>]*>(.*?)</a>`)
 kzIDRE = regexp.MustCompile(`[?&]id=(\d+)`)
)

func parseKinozalResults(tracker TrackerConfig, query SearchQuery, page string) []SearchResult {
 rows:= kzRowRE.FindAllStringSubmatch(page, -1)
 out:= make([]SearchResult, 0)

 for _, row:= range rows {
 link:= kzLinkRE.FindStringSubmatch(row[1])
 if link == nil {
 continue
 }

 id:= ""
 if m:= kzIDRE.FindStringSubmatch(link[1]); m!= nil {
 id = m[1]
 }
 if id == "" {
 continue
 }

 href:= link[1]
 details:= absoluteURL(tracker.BaseURL, href)
 download:= absoluteURL(
 tracker.BaseURL,
 strings.Replace(href, "details.php", "download.php", 1),
 )

 cells:= rtCellRE.FindAllStringSubmatch(row[1], -1)

 result:= SearchResult{
 TrackerID: "kinozal",
 Title: plainHTML(link[2]),
 DetailsURL: details,
 TorrentURL: download,
 GUID: "kinozal:" + id,
 Categories: append([]int(nil), query.Categories...),
 }

 if len(cells) >= 6 {
 result.Size = parseKinozalSize(plainHTML(cells[3][2]))
 result.Seeders = firstInt(plainHTML(cells[4][2]))
 result.Leechers = firstInt(plainHTML(cells[5][2]))
 }

 if len(cells) >= 7 {
 result.PublishDate = parseKinozalDate(plainHTML(cells[6][2]))
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

func parseKinozalSize(s string) int64 {
 s = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(s), ",", "."))
 fields:= strings.Fields(s)
 if len(fields) < 2 {
 return 0
 }

 n, err:= strconv.ParseFloat(fields[0], 64)
 if err!= nil {
 return 0
 }

 var multiplier float64 = 1
 switch fields[1] {
 case "КБ", "KB":
 multiplier = 1 << 10
 case "МБ", "MB":
 multiplier = 1 << 20
 case "ГБ", "GB":
 multiplier = 1 << 30
 case "ТБ", "TB":
 multiplier = 1 << 40
 }

 return int64(n * multiplier)
}

func parseKinozalDate(s string) time.Time {
 now:= time.Now()
 s = strings.TrimSpace(strings.ReplaceAll(s, " в ", " "))
 lower:= strings.ToLower(s)

 if lower == "сейчас" || lower == "now" {
 return now
 }

 for word, days:= range map[string]int{"сегодня": 0, "today": 0, "вчера": -1, "yesterday": -1} {
 if strings.HasPrefix(lower, word) {
 clock:= strings.TrimSpace(s[len(word):])
 if t, err:= time.Parse("15:04", clock); err == nil {
 d:= now.AddDate(0, 0, days)
 return time.Date(d.Year(), d.Month(), d.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
 }
 }
 }

 if t, err:= time.ParseInLocation("02.01.2006 15:04", s, now.Location()); err == nil {
 return t
 }

 return time.Time{}
}
