package converter

import (
	"gtfs-netex-converter/gtfs"
	"testing"
)

func TestNormalizeGTFSClockTime(t *testing.T) {
	testCases := []struct {
		input          string
		expectedTime   string
		expectedOffset int
	}{
		{"00:00:00", "00:00:00", 0},
		{"12:30:45", "12:30:45", 0},
		{"23:59:59", "23:59:59", 0},
		{"24:00:00", "00:00:00", 1},
		{"25:15:30", "01:15:30", 1},
		{"26:00:00", "02:00:00", 1},
		{"48:00:00", "00:00:00", 2},
		{"", "", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			time, offset := normalizeGTFSClockTime(tc.input)
			if time != tc.expectedTime {
				t.Errorf("Expected time %s, got %s", tc.expectedTime, time)
			}
			if offset != tc.expectedOffset {
				t.Errorf("Expected offset %d, got %d", tc.expectedOffset, offset)
			}
		})
	}
}

func TestGetTransportModeFromType(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}
	c := NewConverter(config)

	testCases := []struct {
		routeType    int
		expectedMode string
	}{
		{RouteTypeTram, "tram"},
		{RouteTypeMetro, "metro"},
		{RouteTypeRail, "rail"},
		{RouteTypeBus, "bus"},
		{RouteTypeFerry, "water"},
		{RouteTypeCableTram, "cableway"},
		{RouteTypeAerialLift, "cableway"},
		{RouteTypeFunicular, "funicular"},
		{RouteTypeTrolleybus, "trolleyBus"},
		{RouteTypeMonorail, "rail"},
		{999, "bus"}, // Unknown type defaults to bus
	}

	for _, tc := range testCases {
		t.Run(tc.expectedMode, func(t *testing.T) {
			mode := c.getTransportModeFromType(tc.routeType)
			if mode != tc.expectedMode {
				t.Errorf("Route type %d: expected %s, got %s", tc.routeType, tc.expectedMode, mode)
			}
		})
	}
}

func TestGetVehicleTypeName(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}
	c := NewConverter(config)

	testCases := []struct {
		routeType    int
		expectedName string
	}{
		{RouteTypeTram, "Tram"},
		{RouteTypeMetro, "Metro"},
		{RouteTypeRail, "Rail"},
		{RouteTypeBus, "Bus"},
		{RouteTypeFerry, "Ferry"},
		{RouteTypeCableTram, "Cable Car"},
		{RouteTypeAerialLift, "Gondola"},
		{RouteTypeFunicular, "Funicular"},
		{RouteTypeTrolleybus, "Trolleybus"},
		{RouteTypeMonorail, "Monorail"},
		{999, "Bus"}, // Unknown type defaults to Bus
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			name := c.getVehicleTypeName(tc.routeType)
			if name != tc.expectedName {
				t.Errorf("Route type %d: expected %s, got %s", tc.routeType, tc.expectedName, name)
			}
		})
	}
}

func TestGetRouteType(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}
	c := NewConverter(config)

	// Setup test data
	c.gtfsData = &GTFSData{
		Routes: []gtfs.Route{
			{RouteID: "route1", RouteType: RouteTypeBus},
			{RouteID: "route2", RouteType: RouteTypeTram},
		},
	}

	// Build lookup indices
	c.lookupIndex = BuildLookupIndices(c.gtfsData)

	t.Run("ExistingRoute", func(t *testing.T) {
		routeType := c.getRouteType("route1")
		if routeType != RouteTypeBus {
			t.Errorf("Expected route type %d, got %d", RouteTypeBus, routeType)
		}

		routeType = c.getRouteType("route2")
		if routeType != RouteTypeTram {
			t.Errorf("Expected route type %d, got %d", RouteTypeTram, routeType)
		}
	})

	t.Run("NonexistentRoute", func(t *testing.T) {
		routeType := c.getRouteType("nonexistent")
		if routeType != RouteTypeBus {
			t.Errorf("Expected default route type %d, got %d", RouteTypeBus, routeType)
		}
	})
}

func TestGetDepartureTime(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}
	c := NewConverter(config)

	t.Run("WithStopTimes", func(t *testing.T) {
		stopTimes := []gtfs.StopTime{
			{DepartureTime: "08:30:00"},
			{DepartureTime: "08:35:00"},
		}
		depTime := c.getDepartureTime(stopTimes)
		if depTime != "08:30:00" {
			t.Errorf("Expected 08:30:00, got %s", depTime)
		}
	})

	t.Run("EmptyStopTimes", func(t *testing.T) {
		stopTimes := []gtfs.StopTime{}
		depTime := c.getDepartureTime(stopTimes)
		if depTime != "" {
			t.Errorf("Expected empty string, got %s", depTime)
		}
	})
}

func TestCalculateJourneyDuration(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}
	c := NewConverter(config)

	t.Run("ValidDuration", func(t *testing.T) {
		stopTimes := []gtfs.StopTime{
			{DepartureTime: "08:00:00", ArrivalTime: "08:00:00"},
			{DepartureTime: "08:30:00", ArrivalTime: "08:30:00"},
		}
		duration := c.calculateJourneyDuration(stopTimes)
		if duration != "PT0H30M" {
			t.Errorf("Expected PT0H30M, got %s", duration)
		}
	})

	t.Run("LongDuration", func(t *testing.T) {
		stopTimes := []gtfs.StopTime{
			{DepartureTime: "08:00:00", ArrivalTime: "08:00:00"},
			{DepartureTime: "10:45:00", ArrivalTime: "10:45:00"},
		}
		duration := c.calculateJourneyDuration(stopTimes)
		if duration != "PT2H45M" {
			t.Errorf("Expected PT2H45M, got %s", duration)
		}
	})

	t.Run("EmptyStopTimes", func(t *testing.T) {
		stopTimes := []gtfs.StopTime{}
		duration := c.calculateJourneyDuration(stopTimes)
		if duration != "PT0M" {
			t.Errorf("Expected PT0M, got %s", duration)
		}
	})
}

func TestConstants(t *testing.T) {
	t.Run("RouteTypes", func(t *testing.T) {
		if RouteTypeTram != 0 {
			t.Errorf("RouteTypeTram should be 0, got %d", RouteTypeTram)
		}
		if RouteTypeBus != 3 {
			t.Errorf("RouteTypeBus should be 3, got %d", RouteTypeBus)
		}
		if RouteTypeTrolleybus != 11 {
			t.Errorf("RouteTypeTrolleybus should be 11, got %d", RouteTypeTrolleybus)
		}
	})

	t.Run("NetExVersion", func(t *testing.T) {
		if NetExVersion != "1" {
			t.Errorf("NetExVersion should be '1', got %s", NetExVersion)
		}
	})

	t.Run("TimeConstants", func(t *testing.T) {
		if SecondsPerHour != 3600 {
			t.Errorf("SecondsPerHour should be 3600, got %d", SecondsPerHour)
		}
		if SecondsPerMinute != 60 {
			t.Errorf("SecondsPerMinute should be 60, got %d", SecondsPerMinute)
		}
		if HoursPerDay != 24 {
			t.Errorf("HoursPerDay should be 24, got %d", HoursPerDay)
		}
	})
}
