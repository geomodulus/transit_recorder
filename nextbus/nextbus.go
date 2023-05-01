package nextbus

import (
	"net/url"
	"strconv"
	"time"
)

// VehicleLocations represents one time-scoped update to transit vehicle positions.
type VehicleLocations struct {
	Timestamp LastTime          `json:"lastTime"`
	Locations []VehicleLocation `json:"vehicle"`
}

type LastTime struct {
	Time string `json:"time"`
}

type VehicleLocation struct {
	RouteTag  string `json:"routeTag"`
	DirTag    string `json:"dirTag"`
	VehicleID string `json:"id"`
	Lat       string `json:"lat"`
	Lon       string `json:"lon"`
	// Speed in km/h
	Speed string `json:"speedKmHr"`
	// Age in seconds.
	Age     string `json:"secsSinceReport"`
	Heading string `json:"heading"`
}

// RouteConfigURLforTag takes a routeList URL and a tag and generates a
// new routeConfig URL.
func RouteConfigURLForTag(routeListURL *url.URL, tag string) *url.URL {
	configURL := *routeListURL
	vals := configURL.Query()
	vals.Set("command", "routeConfig")
	vals.Set("r", tag)
	configURL.RawQuery = vals.Encode()
	return &configURL
}

// NextUpdateURL returns the URL for the next update.
func NextUpdateURL(updateURL *url.URL, lastUpdateTime time.Time, routeTag string) *url.URL {
	nextUpdateURL := *updateURL
	vals := nextUpdateURL.Query()
	var t int64
	if lastUpdateTime.IsZero() {
		t = 0
	} else {
		t = lastUpdateTime.UnixNano() / int64(time.Millisecond)
	}
	vals.Set("t", strconv.FormatInt(t, 10))
	vals.Set("r", routeTag)
	nextUpdateURL.RawQuery = vals.Encode()
	return &nextUpdateURL
}
