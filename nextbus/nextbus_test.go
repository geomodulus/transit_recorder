// package nextbus contains functions for interacting with the NextBus API.
package nextbus

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRouteConfigURLForTag(t *testing.T) {
	routeListURL, _ := url.Parse("http://webservices.nextbus.com/service/publicJSONFeed?command=routeList&a=ttc")
	want := "http://webservices.nextbus.com/service/publicJSONFeed?a=ttc&command=routeConfig&r=503"
	got := RouteConfigURLForTag(routeListURL, "503")
	if diff := cmp.Diff(want, got.String()); diff != "" {
		t.Errorf("unexpected URL returned diff:\n%s", diff)
	}
}
