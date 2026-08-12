package main

import (
 "fmt"
 "net/http"
 "net/url"
 "strings"
)

func searchNNMClub(
 client *http.Client,
 tracker TrackerConfig,
 query SearchQuery,
) ([]SearchResult, error) {
 values:= url.Values{
 "f[]": {"-1"},
 "o": {"1"},
 "s": {"2"},
 "tm": {"-1"},
 "shf": {"1"},
 "sha": {"1"},
 "ta": {"-1"},
 "sns": {"-1"},
 "sds": {"4"},
 "nm": {strings.TrimSpace(query.Query)},
 "submit": {"Поиск"},
 }

 body:= formEncodeCP1251(values)
 target:= strings.TrimRight(tracker.BaseURL, "/") +
 "/forum/tracker.php"

 req, err:= http.NewRequest(
 http.MethodPost,
 target,
 strings.NewReader(body),
 )
 if err!= nil {
 return nil, fmt.Errorf("search request: %w", err)
 }

 req.Header.Set("User-Agent", defaultUA)
 req.Header.Set(
 "Content-Type",
 "application/x-www-form-urlencoded; charset=windows-1251",
 )
 req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")
 req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en;q=0.5")

 resp, err:= client.Do(req)
 if err!= nil {
 return nil, fmt.Errorf("search request: %w", err)
 }

 data, err:= readResponseBody(resp)
 if err!= nil {
 return nil, fmt.Errorf("search response: %w", err)
 }

 if resp.StatusCode < 200 || resp.StatusCode >= 400 {
 return nil, fmt.Errorf("search HTTP status: %s", resp.Status)
 }

 page:= decodeRuTrackerText(data)
 lower:= strings.ToLower(page)

 if strings.Contains(lower, "cf-chl-") ||
 strings.Contains(lower, "just a moment") {
 return nil, fmt.Errorf("NNMClub returned Cloudflare challenge")
 }

 return parseNNMClubResults(tracker, query, page), nil
}
