package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/geomodulus/transit_recorder/recording"
)

type VehicleLocation struct {
	DirTag     string  `json:"DirTag"`
	VehicleID  int     `json:"VehicleID"`
	Lat        float64 `json:"Lat"`
	Lon        float64 `json:"Lon"`
	Speed      int     `json:"Speed"`
	TimeOffset int     `json:"TimeOffset"`
}

func ReadVehicleLocations(db *sql.DB, routeTag string, startTime, endTime time.Time) ([]*VehicleLocation, error) {
	timeFormat := "2006-01-02 15:04:05"
	query := `SELECT dir_tag, vehicle_id, latitude, longitude, speed, creation_timestamp FROM vehicle_locations WHERE creation_timestamp BETWEEN ? AND ? AND route_tag = ? ORDER BY creation_timestamp`
	rows, err := db.Query(query, startTime.Format(timeFormat), endTime.Format(timeFormat), routeTag)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []*VehicleLocation{}
	for rows.Next() {
		var record VehicleLocation
		var creationTimestamp time.Time
		err := rows.Scan(&record.DirTag, &record.VehicleID, &record.Lat, &record.Lon, &record.Speed, &creationTimestamp)
		if err != nil {
			return nil, err
		}
		record.TimeOffset = int(creationTimestamp.Sub(startTime).Seconds())
		records = append(records, &record)
	}

	return records, nil
}

type timeValue struct {
	value time.Time
}

// Set is the method to set the flag value, part of the flag.Value interface.
func (t *timeValue) Set(value string) error {
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	t.value = parsedTime
	return nil
}

// String is the method to format the flag's value, part of the flag.Value interface.
func (t *timeValue) String() string {
	return t.value.Format(time.RFC3339)
}

func main() {
	routes := flag.String("routes", "", "comma-separated list of TTC routes to record")

	startTime, endTime := &timeValue{}, &timeValue{}
	flag.Var(startTime, "start", "Time value in RFC3339 format (e.g. 2023-04-30T12:00:00)")
	flag.Var(endTime, "end", "Time value in RFC3339 format (e.g. 2023-04-30T12:00:00)")

	flag.Parse()

	db, err := sql.Open("sqlite3", "./db/recordings.db")
	if err != nil {
		panic(err)
	}

	switch flag.Arg(0) {
	case "record":
		var wg sync.WaitGroup
		cancelCh := make(chan struct{})

		for _, route := range strings.Split(*routes, ",") {
			wg.Add(1)
			go func(route string) {
				fmt.Printf("Recording route %q...\n", route)
				r := recording.New(route)
				if err := r.NextUpdate(db); err != nil {
					fmt.Printf("error fetching update for %q: %v\n", r.RouteTag, err)
					wg.Done()
					return
				}

				// If it worked once, fire it up.
				for {
					select {
					case <-cancelCh:
						wg.Done()
						return
					case <-time.After(15 * time.Second):
						if err := r.NextUpdate(db); err != nil {
							fmt.Printf("error fetching update for %q: %v\n", r.RouteTag, err)
						}
					}
				}
			}(route)
		}
		// Set up a signal listener to listen for SIGINT and SIGTERM signals
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
		// Wait for a signal to stop the program
		sig := <-signalCh
		fmt.Printf("Received signal: %v, stopping ongoing recordings...\n", sig)

		// Send cancel message to all long-running goroutines
		close(cancelCh)

		// Wait for all long-running goroutines to complete
		wg.Wait()

	case "export":
		if startTime.value.IsZero() || endTime.value.IsZero() || *routes == "" {
			fmt.Println("start and end times and at least one route must be specified")
			os.Exit(1)
		}

		for _, route := range strings.Split(*routes, ",") {
			records, err := ReadVehicleLocations(db, route, startTime.value, endTime.value)
			if err != nil {
				panic(err)
			}
			fmt.Println("records:", len(records))
			data, err := json.MarshalIndent(records, "", "  ")
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(fmt.Sprintf("records-%s.json", route), data, 0644)
			if err != nil {
				panic(err)
			}
		}
		fmt.Println("done")
	}
}
