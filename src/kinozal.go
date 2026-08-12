package main

import (
 "fmt"
 "net/http"
 "net/url"
 "regexp"
 "strings"
)

var kinozalSearchRE = regexp.MustCompile(`[^a-zA-Zа-яА-ЯёЁ0-9]+`)

func searchKinozal(
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

 if err:= loginKinozal(client, tracker); err!= nil {
 return nil, err
 }

 params:= url.Values{
 "c": {"0"},
 "s": {normalizeKinozalQuery(query.Query)},
 "g": {"0"},
 "v": {"0"},
 "d": {"0"},
 "w": {"0"},
 "t": {"0"},
 "f": {"0"},
 }

 searchURL:= strings.TrimRight(tracker.BaseURL, "/") +
 "/browse.php?" + params.Encode()

 resp, err:= trackerRequest(client, http.MethodGet, searchURL, nil)
 if err!= nil {
 return nil, fmt.Errorf("search request: %w", err)
 }

 data, err:= readResponseBody(resp)
 if err!= nil {
 return nil, fmt.Errorf("search response: %w", err)
 }

 page:= decodeRuTrackerText(data)

 return parseKinozalResults(tracker, query, page), nil
}

func loginKinozal(client *http.Client, tracker TrackerConfig) error {
 baseURL:= strings.TrimRight(tracker.BaseURL, "/")

 form:= url.Values{
 "username": {tracker.Username},
 "password": {tracker.Password},
 }

 resp, err:= trackerRequest(
 client,
 http.MethodPost,
 baseURL+"/takelogin.php",
 form,
 )
 if err!= nil {
 return fmt.Errorf("login request: %w", err)
 }

 if _, err:= readResponseBody(resp); err!= nil {
 return fmt.Errorf("login response: %w", err)
 }

 resp, err = trackerRequest(
 client,
 http.MethodGet,
 baseURL+"/my.php",
 nil,
 )
 if err!= nil {
 return fmt.Errorf("login test: %w", err)
 }

 data, err:= readResponseBody(resp)
 if err!= nil {
 return fmt.Errorf("login test response: %w", err)
 }

 page:= decodeRuTrackerText(data)
 if!strings.Contains(page, "logout.php?hash4u=") {
 return fmt.Errorf("Kinozal authentication failed")
 }

 return nil
}

func normalizeKinozalQuery(query string) string {
 query = strings.TrimSpace(query)
 query = kinozalSearchRE.ReplaceAllString(query, " ")
 return strings.Join(strings.Fields(query), " ")
}
