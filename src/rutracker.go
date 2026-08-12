package main

import (
 "fmt"
 "net/http"
 "net/url"
 "regexp"
 "strings"
)

var rtSearchRE = regexp.MustCompile(`[^a-zA-Zа-яА-ЯёЁ0-9]+`)

func searchRuTracker(
 baseClient *http.Client,
 tracker TrackerConfig,
 query SearchQuery,
) ([]SearchResult, error) {
 if strings.TrimSpace(tracker.Username) == "" ||
 strings.TrimSpace(tracker.Password) == "" {
 return nil, fmt.Errorf("username and password are required")
 }

 client, err:= newSessionClient(baseClient)
 if err!= nil {
 return nil, err
 }

 if err:= loginRuTracker(client, tracker); err!= nil {
 return nil, err
 }

 searchURL:= strings.TrimRight(tracker.BaseURL, "/") +
 "/forum/tracker.php?" + url.Values{
 "nm": {normalizeRuTrackerQuery(query.Query)},
 }.Encode()

 resp, err:= trackerRequest(
 client,
 http.MethodGet,
 searchURL,
 nil,
 )
 if err!= nil {
 return nil, fmt.Errorf("search request: %w", err)
 }

 data, err:= readResponseBody(resp)
 if err!= nil {
 return nil, fmt.Errorf("search response: %w", err)
 }

 page:= decodeRuTrackerText(data)

 if!strings.Contains(page, `id="logged-in-username"`) &&
!strings.Contains(page, `id='logged-in-username'`) {
 return nil, fmt.Errorf("RuTracker session is not authenticated")
 }

 return parseRuTrackerResults(tracker, query, page), nil
}

func loginRuTracker(
 client *http.Client,
 tracker TrackerConfig,
) error {
 loginURL:= strings.TrimRight(tracker.BaseURL, "/") +
 "/forum/login.php"

 resp, err:= trackerRequest(
 client,
 http.MethodGet,
 loginURL,
 nil,
 )
 if err!= nil {
 return fmt.Errorf("login page: %w", err)
 }

 data, err:= readResponseBody(resp)
 if err!= nil {
 return fmt.Errorf("login page response: %w", err)
 }

 page:= decodeRuTrackerText(data)

 if strings.Contains(page, "static.rutracker.cc/captcha/") ||
 strings.Contains(page, `name="cap_sid"`) {
 return fmt.Errorf("RuTracker requested CAPTCHA")
 }

 form:= url.Values{
 "login_username": {tracker.Username},
 "login_password": {tracker.Password},
 "login": {"Login"},
 }

 resp, err = trackerRequest(
 client,
 http.MethodPost,
 loginURL,
 form,
 )
 if err!= nil {
 return fmt.Errorf("login request: %w", err)
 }

 data, err = readResponseBody(resp)
 if err!= nil {
 return fmt.Errorf("login response: %w", err)
 }

 page = decodeRuTrackerText(data)

 if!strings.Contains(page, `id="logged-in-username"`) &&
!strings.Contains(page, `id='logged-in-username'`) {
 return fmt.Errorf("RuTracker authentication failed")
 }

 return nil
}

func normalizeRuTrackerQuery(query string) string {
 query = strings.TrimSpace(query)
 if query == "" {
 return ""
 }

 query = rtSearchRE.ReplaceAllString(query, "%")
 return strings.Trim(query, "%")
}
