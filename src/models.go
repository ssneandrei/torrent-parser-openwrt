package main

import (
 "net/http"
 "time"
)

type SearchQuery struct {
 Query string
 Categories []int
 Limit int
 Offset int
}

type SearchResult struct {
 TrackerID string
 Title string
 DetailsURL string
 TorrentURL string
 MagnetURL string
 GUID string
 Size int64
 Seeders int
 Leechers int
 PublishDate time.Time
 Categories []int
}

type TrackerSearchFunc func(
 client *http.Client,
 tracker TrackerConfig,
 query SearchQuery,
) ([]SearchResult, error)

type TrackerAdapter struct {
 ID string
 Search TrackerSearchFunc
}
