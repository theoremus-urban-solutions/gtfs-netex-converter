package converter

import (
	"strings"
	"testing"
)

func TestIDGenerator(t *testing.T) {
	gen := NewIDGenerator("BG")

	t.Run("GenerateID", func(t *testing.T) {
		id := gen.GenerateID("TestType", "test123")
		expected := "BG::TestType:test123::"
		if id != expected {
			t.Errorf("Expected %s, got %s", expected, id)
		}
	})

	t.Run("GenerateStopPlaceID", func(t *testing.T) {
		id := gen.GenerateStopPlaceID("stop1")
		if !strings.Contains(id, "BG::StopPlace:stop_place_stop1::") {
			t.Errorf("Unexpected stop place ID format: %s", id)
		}
	})

	t.Run("GenerateServiceJourneyID", func(t *testing.T) {
		id := gen.GenerateServiceJourneyID("trip123")
		if !strings.Contains(id, "BG::ServiceJourney:TRIP_trip123::") {
			t.Errorf("Unexpected service journey ID format: %s", id)
		}
	})

	t.Run("DifferentParticipantRef", func(t *testing.T) {
		gen2 := NewIDGenerator("FR")
		id := gen2.GenerateStopPlaceID("stop1")
		if !strings.HasPrefix(id, "FR::") {
			t.Errorf("Expected ID to start with FR::, got %s", id)
		}
	})
}

func TestIDGeneratorConsistency(t *testing.T) {
	gen := NewIDGenerator("BG")

	// Generate same ID twice
	id1 := gen.GenerateStopPlaceID("stop1")
	id2 := gen.GenerateStopPlaceID("stop1")

	if id1 != id2 {
		t.Errorf("Same input should generate same ID. Got %s and %s", id1, id2)
	}
}

func TestIDGeneratorUniqueness(t *testing.T) {
	gen := NewIDGenerator("BG")

	// Different inputs should generate different IDs
	id1 := gen.GenerateStopPlaceID("stop1")
	id2 := gen.GenerateStopPlaceID("stop2")

	if id1 == id2 {
		t.Errorf("Different inputs should generate different IDs")
	}
}

func TestAllIDGeneratorMethods(t *testing.T) {
	gen := NewIDGenerator("TEST")

	testCases := []struct {
		name     string
		method   func() string
		contains string
	}{
		{"StopPlace", func() string { return gen.GenerateStopPlaceID("s1") }, "TEST::StopPlace:"},
		{"Quay", func() string { return gen.GenerateQuayID("s1") }, "TEST::Quay:"},
		{"ScheduledStopPoint", func() string { return gen.GenerateScheduledStopPointID("s1") }, "TEST::ScheduledStopPoint:"},
		{"Route", func() string { return gen.GenerateRouteID("r1") }, "TEST::Route:"},
		{"Line", func() string { return gen.GenerateLineID("r1") }, "TEST::Line:"},
		{"ServiceJourney", func() string { return gen.GenerateServiceJourneyID("t1") }, "TEST::ServiceJourney:"},
		{"JourneyPattern", func() string { return gen.GenerateJourneyPatternID("p1") }, "TEST::JourneyPattern:"},
		{"DayType", func() string { return gen.GenerateDayTypeID("d1") }, "TEST::DayType:"},
		{"OperatingPeriod", func() string { return gen.GenerateOperatingPeriodID("o1") }, "TEST::OperatingPeriod:"},
		{"VehicleType", func() string { return gen.GenerateVehicleTypeID(3) }, "TEST::VehicleType:"},
		{"Direction", func() string { return gen.GenerateDirectionID("0") }, "TEST::Direction:"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.method()
			if !strings.Contains(id, tc.contains) {
				t.Errorf("Expected ID to contain %s, got %s", tc.contains, id)
			}
			if !strings.HasPrefix(id, "TEST::") {
				t.Errorf("Expected ID to start with TEST::, got %s", id)
			}
			if !strings.HasSuffix(id, "::") {
				t.Errorf("Expected ID to end with ::, got %s", id)
			}
		})
	}
}
