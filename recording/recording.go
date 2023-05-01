// Package recording contains code for recording vehicle location data.
package recording

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/geomodulus/transit_recorder/nextbus"
)

const (
	TTCVehicleLocationsBaseURL = "https://webservices.umoiq.com/service/publicJSONFeed?command=vehicleLocations&a=ttc" //&t=0&r=510"
)

type Recording struct {
	RouteTag string
	lastTime time.Time
}

func New(routeTag string) *Recording {
	return &Recording{
		RouteTag: routeTag,
		lastTime: time.Time{},
	}
}

// NextUpdate requests the next vehicle location update from the TTC API.
func (r *Recording) NextUpdate(db *sql.DB) error {
	baseURL, err := url.Parse(TTCVehicleLocationsBaseURL)
	if err != nil {
		return fmt.Errorf("error parsing base URL: %v", err)
	}
	nextUpdateURL := nextbus.NextUpdateURL(baseURL, r.lastTime, r.RouteTag)

	//	fmt.Println("-------------------------------")
	//	fmt.Println("Requesting", nextUpdateURL.String())
	resp, err := http.Get(nextUpdateURL.String())
	if err != nil {
		return fmt.Errorf("error fetching next update URL: %v", err)
	}
	defer resp.Body.Close()

	r.lastTime = time.Now()

	var vehicleLocations nextbus.VehicleLocations
	if err := json.NewDecoder(resp.Body).Decode(&vehicleLocations); err != nil {
		return fmt.Errorf("error decoding response body: %v", err)
	}

	stmt, err := db.Prepare("INSERT INTO vehicle_locations (route_tag, dir_tag, vehicle_id, latitude, longitude, speed, age, heading) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("error preparing insert statement: %v", err)
	}
	defer stmt.Close()

	for _, location := range vehicleLocations.Locations {
		//		fmt.Printf("%+v\n", location)
		_, err = stmt.Exec(location.RouteTag, location.DirTag, location.VehicleID, location.Lat, location.Lon, location.Speed, location.Age, location.Heading)
		if err != nil {
			return fmt.Errorf("error inserting vehicle location (VehicleID: %s): %v", location.VehicleID, err)
		}
	}

	fmt.Println(r.RouteTag, "updated with", len(vehicleLocations.Locations), "active vehicles")

	return nil
}
