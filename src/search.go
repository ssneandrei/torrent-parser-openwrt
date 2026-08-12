package main

import (
 "fmt"
 "sync"
)

type trackerSearchResult struct {
 trackerID string
 results []SearchResult
 err error
}

func (a *App) searchAll(query SearchQuery) ([]SearchResult, []error) {
 adapters:= map[string]TrackerAdapter{
 "rutracker": {
 ID: "rutracker",
 Search: searchRuTracker,
 },
 "kinozal": {
 ID: "kinozal",
 Search: searchKinozal,
 },
 "nnmclub": {
 ID: "nnmclub",
 Search: searchNNMClub,
 },
 }

 ch:= make(chan trackerSearchResult, len(adapters))
 var wg sync.WaitGroup

 for id, adapter:= range adapters {
 tracker, ok:= a.cfg.Trackers[id]
 if!ok ||!tracker.Enabled {
 continue
 }

 wg.Add(1)

 go func(adapter TrackerAdapter, tracker TrackerConfig) {
 defer wg.Done()

 results, err:= adapter.Search(
 a.client,
 tracker,
 query,
 )

 ch <- trackerSearchResult{
 trackerID: tracker.ID,
 results: results,
 err: err,
 }
 }(adapter, tracker)
 }

 go func() {
 wg.Wait()
 close(ch)
 }()

 combined:= make([]SearchResult, 0)
 errs:= make([]error, 0)

 for result:= range ch {
 if result.err!= nil {
 errs = append(
 errs,
 fmt.Errorf("%s: %w", result.trackerID, result.err),
 )
 continue
 }

 combined = append(combined, result.results...)
 }

 return deduplicateResults(combined), errs
}

func deduplicateResults(results []SearchResult) []SearchResult {
 seen:= make(map[string]struct{})
 out:= make([]SearchResult, 0, len(results))

 for _, result:= range results {
 key:= result.GUID

 if key == "" {
 key = result.TrackerID + ":" + result.DetailsURL
 }

 if _, exists:= seen[key]; exists {
 continue
 }

 seen[key] = struct{}{}
 out = append(out, result)
 }

 return out
}
