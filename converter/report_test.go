package converter

import (
	"encoding/json"
	"testing"
	"time"

	"gtfs-netex-converter/gtfs"
)

func TestGenerateReport(t *testing.T) {
	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "bg",
	}

	c := NewConverter(config)
	c.inputDir = "./test_data"
	c.startTime = time.Now()
	c.endTime = c.startTime.Add(2 * time.Second)

	// Setup test data
	c.gtfsData = &GTFSData{
		Agencies:       []gtfs.Agency{{AgencyID: "1", AgencyName: "Test Agency"}},
		Stops:          make([]gtfs.Stop, 10),
		Routes:         make([]gtfs.Route, 5),
		Trips:          make([]gtfs.Trip, 50),
		StopTimes:      make([]gtfs.StopTime, 200),
		CalendarDates:  make([]gtfs.CalendarDate, 30),
		Shapes:         make([]gtfs.Shape, 100),
		Transfers:      make([]gtfs.Transfer, 15),
		FareAttributes: make([]gtfs.FareAttribute, 2),
		Pathways:       make([]gtfs.Pathway, 3),
	}

	report := c.GenerateReport()

	t.Run("Summary", func(t *testing.T) {
		if report.Summary.Status != "success" {
			t.Errorf("Expected status 'success', got %s", report.Summary.Status)
		}

		if report.Summary.ParticipantRef != "TEST" {
			t.Errorf("Expected participant 'TEST', got %s", report.Summary.ParticipantRef)
		}

		if report.Summary.InputDirectory != "./test_data" {
			t.Errorf("Expected input directory './test_data', got %s", report.Summary.InputDirectory)
		}

		if report.Summary.DurationSeconds <= 0 {
			t.Errorf("Expected positive duration, got %f", report.Summary.DurationSeconds)
		}
	})

	t.Run("InputStatistics", func(t *testing.T) {
		if report.InputStatistics.Agencies != 1 {
			t.Errorf("Expected 1 agency, got %d", report.InputStatistics.Agencies)
		}

		if report.InputStatistics.Stops != 10 {
			t.Errorf("Expected 10 stops, got %d", report.InputStatistics.Stops)
		}

		if report.InputStatistics.Routes != 5 {
			t.Errorf("Expected 5 routes, got %d", report.InputStatistics.Routes)
		}

		if report.InputStatistics.Trips != 50 {
			t.Errorf("Expected 50 trips, got %d", report.InputStatistics.Trips)
		}

		if report.InputStatistics.StopTimes != 200 {
			t.Errorf("Expected 200 stop times, got %d", report.InputStatistics.StopTimes)
		}

		totalRecords := 1 + 10 + 5 + 50 + 200 + 30 + 100 + 15 + 2 + 3
		if report.InputStatistics.TotalRecords != totalRecords {
			t.Errorf("Expected %d total records, got %d", totalRecords, report.InputStatistics.TotalRecords)
		}
	})

	t.Run("OutputStatistics", func(t *testing.T) {
		if report.OutputStatistics.Authorities != 1 {
			t.Errorf("Expected 1 authority, got %d", report.OutputStatistics.Authorities)
		}

		if report.OutputStatistics.StopPlaces != 10 {
			t.Errorf("Expected 10 stop places, got %d", report.OutputStatistics.StopPlaces)
		}

		if report.OutputStatistics.Lines != 5 {
			t.Errorf("Expected 5 lines, got %d", report.OutputStatistics.Lines)
		}

		if report.OutputStatistics.ServiceJourneys != 50 {
			t.Errorf("Expected 50 service journeys, got %d", report.OutputStatistics.ServiceJourneys)
		}
	})

	t.Run("EntityMapping", func(t *testing.T) {
		agencyMapping, exists := report.EntityMapping["Agency->Authority"]
		if !exists {
			t.Fatal("Agency->Authority mapping not found")
		}

		if agencyMapping.InputCount != 1 {
			t.Errorf("Expected 1 input agency, got %d", agencyMapping.InputCount)
		}

		if agencyMapping.OutputCount != 1 {
			t.Errorf("Expected 1 output authority, got %d", agencyMapping.OutputCount)
		}

		if agencyMapping.SuccessRate != 100.0 {
			t.Errorf("Expected 100%% success rate, got %f", agencyMapping.SuccessRate)
		}
	})

	t.Run("Warnings", func(t *testing.T) {
		// Should have warnings for shapes and fares
		if len(report.Warnings) == 0 {
			t.Error("Expected warnings to be present")
		}

		hasShapesWarning := false
		hasFaresWarning := false

		for _, warning := range report.Warnings {
			if warning.Type == "feature_not_implemented" && warning.Count == 100 {
				hasShapesWarning = true
			}
			if warning.Type == "feature_not_enabled" && warning.Count == 2 {
				hasFaresWarning = true
			}
		}

		if !hasShapesWarning {
			t.Error("Expected shapes warning")
		}

		if !hasFaresWarning {
			t.Error("Expected fares warning")
		}
	})
}

func TestReportJSONSerialization(t *testing.T) {
	report := &ConversionReport{
		Summary: ConversionSummary{
			Status:          "success",
			StartTime:       time.Now(),
			EndTime:         time.Now().Add(2 * time.Second),
			DurationSeconds: 2.0,
			ParticipantRef:  "TEST",
			InputDirectory:  "./test",
			OutputFile:      "./output.xml",
		},
		InputStatistics: InputStatistics{
			FilesProcessed: 5,
			TotalRecords:   100,
			Agencies:       1,
			Stops:          10,
			Routes:         5,
			Trips:          20,
			StopTimes:      64,
		},
		Performance: PerformanceMetrics{
			TotalDurationSeconds: 2.0,
			RecordsPerSecond:     50.0,
		},
		Warnings: []Warning{},
		Errors:   []ConversionError{},
	}

	t.Run("ToJSON", func(t *testing.T) {
		jsonData, err := report.ToJSON()
		if err != nil {
			t.Fatalf("Failed to convert to JSON: %v", err)
		}

		if len(jsonData) == 0 {
			t.Error("Expected non-empty JSON output")
		}

		// Verify it's valid JSON by unmarshaling
		var decoded map[string]interface{}
		if err := json.Unmarshal(jsonData, &decoded); err != nil {
			t.Fatalf("Generated invalid JSON: %v", err)
		}

		// Check key fields exist
		if _, exists := decoded["summary"]; !exists {
			t.Error("Missing 'summary' in JSON")
		}

		if _, exists := decoded["input_statistics"]; !exists {
			t.Error("Missing 'input_statistics' in JSON")
		}

		if _, exists := decoded["performance"]; !exists {
			t.Error("Missing 'performance' in JSON")
		}
	})

	t.Run("ToJSONCompact", func(t *testing.T) {
		jsonData, err := report.ToJSONCompact()
		if err != nil {
			t.Fatalf("Failed to convert to compact JSON: %v", err)
		}

		if len(jsonData) == 0 {
			t.Error("Expected non-empty JSON output")
		}

		// Compact JSON should be smaller (no indentation)
		prettyJSON, _ := report.ToJSON()
		if len(jsonData) >= len(prettyJSON) {
			t.Error("Compact JSON should be smaller than pretty JSON")
		}
	})
}

func TestReportPerformanceMetrics(t *testing.T) {
	config := &Config{
		ParticipantRef: "TEST",
	}

	c := NewConverter(config)
	c.startTime = time.Now()
	c.endTime = c.startTime.Add(5 * time.Second)

	c.gtfsData = &GTFSData{
		Agencies:  []gtfs.Agency{{AgencyID: "1"}},
		Stops:     make([]gtfs.Stop, 1000),
		Routes:    make([]gtfs.Route, 100),
		Trips:     make([]gtfs.Trip, 5000),
		StopTimes: make([]gtfs.StopTime, 50000),
	}

	report := c.GenerateReport()

	t.Run("CalculatesRecordsPerSecond", func(t *testing.T) {
		expectedRecords := 1 + 1000 + 100 + 5000 + 50000
		expectedRate := float64(expectedRecords) / 5.0

		if report.Performance.RecordsPerSecond < expectedRate*0.9 ||
			report.Performance.RecordsPerSecond > expectedRate*1.1 {
			t.Errorf("Expected rate around %f, got %f",
				expectedRate, report.Performance.RecordsPerSecond)
		}
	})

	t.Run("DurationMatches", func(t *testing.T) {
		if report.Performance.TotalDurationSeconds < 4.9 ||
			report.Performance.TotalDurationSeconds > 5.1 {
			t.Errorf("Expected duration around 5s, got %f",
				report.Performance.TotalDurationSeconds)
		}
	})
}

func TestReportWithErrors(t *testing.T) {
	config := &Config{
		ParticipantRef: "TEST",
	}

	c := NewConverter(config)
	c.gtfsData = &GTFSData{}
	c.startTime = time.Now()
	c.endTime = time.Now().Add(1 * time.Second)

	report := c.GenerateReport()

	t.Run("EmptyDataset", func(t *testing.T) {
		if report.InputStatistics.TotalRecords != 0 {
			t.Errorf("Expected 0 records for empty dataset, got %d",
				report.InputStatistics.TotalRecords)
		}

		if report.InputStatistics.FilesProcessed != 0 {
			t.Errorf("Expected 0 files processed, got %d",
				report.InputStatistics.FilesProcessed)
		}
	})
}

func TestEntityMappingStats(t *testing.T) {
	config := &Config{
		ParticipantRef: "TEST",
	}

	c := NewConverter(config)
	c.startTime = time.Now()
	c.endTime = c.startTime.Add(1 * time.Second)

	c.gtfsData = &GTFSData{
		Agencies: []gtfs.Agency{
			{AgencyID: "1"},
			{AgencyID: "2"},
		},
		Stops: make([]gtfs.Stop, 500),
		Routes: []gtfs.Route{
			{RouteID: "R1"},
			{RouteID: "R2"},
			{RouteID: "R3"},
		},
		Trips: make([]gtfs.Trip, 1000),
	}

	report := c.GenerateReport()

	t.Run("AllMappingsPresent", func(t *testing.T) {
		requiredMappings := []string{
			"Agency->Authority",
			"Stop->StopPlace",
			"Route->Line",
			"Trip->ServiceJourney",
		}

		for _, mapping := range requiredMappings {
			if _, exists := report.EntityMapping[mapping]; !exists {
				t.Errorf("Missing required mapping: %s", mapping)
			}
		}
	})

	t.Run("CorrectCounts", func(t *testing.T) {
		agencyMap := report.EntityMapping["Agency->Authority"]
		if agencyMap.InputCount != 2 {
			t.Errorf("Expected 2 agencies, got %d", agencyMap.InputCount)
		}

		stopMap := report.EntityMapping["Stop->StopPlace"]
		if stopMap.InputCount != 500 {
			t.Errorf("Expected 500 stops, got %d", stopMap.InputCount)
		}

		routeMap := report.EntityMapping["Route->Line"]
		if routeMap.InputCount != 3 {
			t.Errorf("Expected 3 routes, got %d", routeMap.InputCount)
		}

		tripMap := report.EntityMapping["Trip->ServiceJourney"]
		if tripMap.InputCount != 1000 {
			t.Errorf("Expected 1000 trips, got %d", tripMap.InputCount)
		}
	})
}
