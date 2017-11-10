package goesi

import (
	"github.com/Jeffail/gabs"
	"net/http"
	"time"
)

// A CacheEntry is a single response from ESI
type CacheEntry struct {
	Data    *gabs.Container
	Expires time.Time
}

// A Cache is a map that stores GET responses from ESI.
// This cache is for for responses to GET requests only - POST
// requests are not cached, as the responses are likely determined
// by what is sent to ESI.
type Cache map[string]CacheEntry

// get returns an entry from the map (if it exists and is not expired).
// If the entry is present but expired, it is removed from the map.
func (c *Cache) get(u string) *gabs.Container {
	entry, ok := (*c)[u]
	if !ok {
		log.Debug("No entry in cache for URL '%s'", u)
		return nil
	}
	// check expiration
	log.Debug("Checking expiration value")
	if entry.Expires.Before(time.Now().UTC()) {
		// removed the expired data from the cache
		log.Debug("Data in cache is expired; removing from cache")
		delete(*c, u)
		return nil
	}
	log.Debug("Returning non-expired cached data")
	return entry.Data
}

// set puts the url and its data into the cache
func (c *Cache) set(u string, d *gabs.Container, h http.Header) error {
	expires, err := getExpiration(h.Get("Expires"))
	log.Debug("Storing url in cache, '%s', expires '%s'", u, expires)
	if err != nil {
		return err
	}
	entry := CacheEntry{d, expires}
	(*c)[u] = entry
	return nil
}

// getExpiration parses the expiration time from the ESI response headers
func getExpiration(s string) (time.Time, error) {
	parseFormat := "Mon, 02 Jan 2006 15:04:05 MST"
	t, err := time.Parse(parseFormat, s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
