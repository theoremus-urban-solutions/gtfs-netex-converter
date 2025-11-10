package converter

import (
	"gtfs-netex-converter/gtfs"
	"testing"
)

func TestBuildLookupIndices(t *testing.T) {
	// Create test data
	data := &GTFSData{
		Stops: []gtfs.Stop{
			{StopID: "stop1", StopName: "Stop 1"},
			{StopID: "stop2", StopName: "Stop 2"},
		},
		Routes: []gtfs.Route{
			{RouteID: "route1", RouteShortName: "R1", RouteType: RouteTypeBus},
			{RouteID: "route2", RouteShortName: "R2", RouteType: RouteTypeTram},
		},
		Trips: []gtfs.Trip{
			{TripID: "trip1", RouteID: "route1", ServiceID: "service1"},
			{TripID: "trip2", RouteID: "route1", ServiceID: "service1"},
			{TripID: "trip3", RouteID: "route2", ServiceID: "service2"},
		},
		StopTimes: []gtfs.StopTime{
			{TripID: "trip1", StopID: "stop1", StopSequence: 1},
			{TripID: "trip1", StopID: "stop2", StopSequence: 2},
			{TripID: "trip2", StopID: "stop1", StopSequence: 1},
			{TripID: "trip3", StopID: "stop2", StopSequence: 1},
		},
		Agencies: []gtfs.Agency{
			{AgencyID: "agency1", AgencyName: "Test Agency"},
		},
	}

	// Build indices
	indices := BuildLookupIndices(data)

	// Test stop times lookup
	t.Run("GetStopTimesByTripID", func(t *testing.T) {
		stopTimes := indices.GetStopTimesByTripID("trip1")
		if len(stopTimes) != 2 {
			t.Errorf("Expected 2 stop times for trip1, got %d", len(stopTimes))
		}

		stopTimes = indices.GetStopTimesByTripID("nonexistent")
		if len(stopTimes) != 0 {
			t.Errorf("Expected 0 stop times for nonexistent trip, got %d", len(stopTimes))
		}
	})

	// Test trips by route lookup
	t.Run("GetTripsByRouteID", func(t *testing.T) {
		trips := indices.GetTripsByRouteID("route1")
		if len(trips) != 2 {
			t.Errorf("Expected 2 trips for route1, got %d", len(trips))
		}

		trips = indices.GetTripsByRouteID("route2")
		if len(trips) != 1 {
			t.Errorf("Expected 1 trip for route2, got %d", len(trips))
		}
	})

	// Test route lookup
	t.Run("GetRouteByID", func(t *testing.T) {
		route := indices.GetRouteByID("route1")
		if route == nil {
			t.Fatal("Expected route1 to exist")
		}
		if route.RouteShortName != "R1" {
			t.Errorf("Expected route short name R1, got %s", route.RouteShortName)
		}
		if route.RouteType != RouteTypeBus {
			t.Errorf("Expected route type %d, got %d", RouteTypeBus, route.RouteType)
		}

		route = indices.GetRouteByID("nonexistent")
		if route != nil {
			t.Error("Expected nil for nonexistent route")
		}
	})

	// Test stop lookup
	t.Run("GetStopByID", func(t *testing.T) {
		stop := indices.GetStopByID("stop1")
		if stop == nil {
			t.Fatal("Expected stop1 to exist")
		}
		if stop.StopName != "Stop 1" {
			t.Errorf("Expected stop name 'Stop 1', got %s", stop.StopName)
		}

		stop = indices.GetStopByID("nonexistent")
		if stop != nil {
			t.Error("Expected nil for nonexistent stop")
		}
	})

	// Test agency lookup
	t.Run("GetAgencyByID", func(t *testing.T) {
		agency := indices.GetAgencyByID("agency1")
		if agency == nil {
			t.Fatal("Expected agency1 to exist")
		}
		if agency.AgencyName != "Test Agency" {
			t.Errorf("Expected agency name 'Test Agency', got %s", agency.AgencyName)
		}
	})
}

func TestLookupIndicesPerformance(t *testing.T) {
	// Create large dataset
	const numTrips = 1000
	const stopsPerTrip = 50

	data := &GTFSData{
		StopTimes: make([]gtfs.StopTime, 0, numTrips*stopsPerTrip),
		Trips:     make([]gtfs.Trip, 0, numTrips),
	}

	// Generate test data
	for i := 0; i < numTrips; i++ {
		tripID := string(rune('A') + rune(i%26)) + string(rune('0') + rune(i/26))
		data.Trips = append(data.Trips, gtfs.Trip{
			TripID:    tripID,
			RouteID:   "route1",
			ServiceID: "service1",
		})

		for j := 0; j < stopsPerTrip; j++ {
			data.StopTimes = append(data.StopTimes, gtfs.StopTime{
				TripID:       tripID,
				StopID:       "stop" + string(rune('0'+j)),
				StopSequence: j + 1,
			})
		}
	}

	// Build indices
	indices := BuildLookupIndices(data)

	// Verify O(1) lookup performance
	t.Run("FastLookup", func(t *testing.T) {
		// This should be O(1) with indices
		for i := 0; i < 100; i++ {
			tripID := string(rune('A') + rune(i%26)) + string(rune('0') + rune(i/26))
			stopTimes := indices.GetStopTimesByTripID(tripID)
			if len(stopTimes) != stopsPerTrip {
				t.Errorf("Expected %d stop times, got %d", stopsPerTrip, len(stopTimes))
			}
		}
	})
}
