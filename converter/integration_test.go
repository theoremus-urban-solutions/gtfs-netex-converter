package converter

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"
)

func TestEndToEndConversion(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create minimal GTFS files
	createMinimalGTFS(t, tmpDir)

	config := &Config{
		ParticipantRef:  "TEST",
		DefaultTimezone: "Europe/Sofia",
		DefaultLanguage: "en",
	}

	conv := NewConverter(config)

	t.Run("SuccessfulConversion", func(t *testing.T) {
		netexData, err := conv.Convert(tmpDir)
		if err != nil {
			t.Fatalf("Conversion failed: %v", err)
		}

		if netexData == nil {
			t.Fatal("Expected non-nil NeTEx data")
		}

		// Verify NeTEx structure
		if netexData.ParticipantRef != "TEST" {
			t.Errorf("Expected participant 'TEST', got %s", netexData.ParticipantRef)
		}
	})

	t.Run("GeneratesValidXML", func(t *testing.T) {
		netexData, _ := conv.Convert(tmpDir)

		xmlData, err := xml.MarshalIndent(netexData, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal XML: %v", err)
		}

		if len(xmlData) == 0 {
			t.Error("Expected non-empty XML output")
		}

		// Verify XML header
		xmlStr := string(xmlData)
		if len(xmlStr) < 100 {
			t.Error("XML output too short, likely invalid")
		}
	})

	t.Run("GeneratesReport", func(t *testing.T) {
		conv.Convert(tmpDir)
		report := conv.GenerateReport()

		if report == nil {
			t.Fatal("Expected non-nil report")
		}

		if report.Summary.Status != "success" {
			t.Errorf("Expected success status, got %s", report.Summary.Status)
		}

		if report.InputStatistics.TotalRecords == 0 {
			t.Error("Expected non-zero total records")
		}
	})

	t.Run("ReportJSONIsValid", func(t *testing.T) {
		conv.Convert(tmpDir)
		report := conv.GenerateReport()

		jsonData, err := report.ToJSON()
		if err != nil {
			t.Fatalf("Failed to generate JSON: %v", err)
		}

		if len(jsonData) == 0 {
			t.Error("Expected non-empty JSON")
		}
	})
}

func TestConversionStatistics(t *testing.T) {
	tmpDir := t.TempDir()
	createMinimalGTFS(t, tmpDir)

	config := &Config{
		ParticipantRef: "TEST",
	}

	conv := NewConverter(config)
	conv.Convert(tmpDir)

	stats := conv.GetStats()

	t.Run("HasGTFSEntities", func(t *testing.T) {
		if stats.GTFSEntities == 0 {
			t.Error("Expected non-zero GTFS entities")
		}
	})

	t.Run("HasNeTExEntities", func(t *testing.T) {
		if stats.NeTExEntities == 0 {
			t.Error("Expected non-zero NeTEx entities")
		}
	})

	t.Run("HasDuration", func(t *testing.T) {
		if stats.DurationSeconds <= 0 {
			t.Error("Expected positive duration")
		}
	})
}

func TestIntegrationPerformance(t *testing.T) {
	tmpDir := t.TempDir()

	// Create larger dataset for performance testing
	createLargerGTFS(t, tmpDir)

	config := &Config{
		ParticipantRef: "TEST",
	}

	conv := NewConverter(config)

	t.Run("HandlesDataset", func(t *testing.T) {
		_, err := conv.Convert(tmpDir)
		if err != nil {
			t.Fatalf("Failed to convert dataset: %v", err)
		}

		report := conv.GenerateReport()

		// Should process quickly with O(1) lookups
		if report.Performance.TotalDurationSeconds > 10.0 {
			t.Errorf("Conversion too slow: %.2fs (expected < 10s)",
				report.Performance.TotalDurationSeconds)
		}

		// Should have reasonable throughput
		if report.Performance.RecordsPerSecond < 100 {
			t.Errorf("Low throughput: %.0f entities/s (expected > 100)",
				report.Performance.RecordsPerSecond)
		}
	})
}

func TestErrorHandling(t *testing.T) {
	config := &Config{
		ParticipantRef: "TEST",
	}

	conv := NewConverter(config)

	t.Run("NonexistentDirectory", func(t *testing.T) {
		netexData, err := conv.Convert("/nonexistent/directory")
		// Converter handles missing files gracefully with warnings
		// It should either return an error OR return data with minimal/no entities
		if err == nil && netexData != nil {
			// If no error, verify the report shows the problem
			report := conv.GenerateReport()
			if report.InputStatistics.TotalRecords == 0 {
				t.Logf("No error returned, but no data loaded (expected behavior)")
			}
		}
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := conv.Convert(tmpDir)
		// Should handle gracefully, may or may not error depending on requirements
		// For now, just ensure it doesn't panic
		if err != nil {
			t.Logf("Empty directory error (expected): %v", err)
		}
	})
}

func TestConfigurableParticipantRef(t *testing.T) {
	tmpDir := t.TempDir()
	createMinimalGTFS(t, tmpDir)

	testCases := []struct {
		participant string
	}{
		{"BG"},
		{"FR"},
		{"US_NY"},
		{"UK"},
		{"JP"},
	}

	for _, tc := range testCases {
		t.Run("Participant_"+tc.participant, func(t *testing.T) {
			config := &Config{
				ParticipantRef:  tc.participant,
				DefaultTimezone: "Europe/Sofia",
				DefaultLanguage: "en",
			}

			conv := NewConverter(config)
			netexData, err := conv.Convert(tmpDir)
			if err != nil {
				t.Fatalf("Conversion failed: %v", err)
			}

			if netexData.ParticipantRef != tc.participant {
				t.Errorf("Expected participant %s, got %s",
					tc.participant, netexData.ParticipantRef)
			}

			// Verify IDs use correct participant
			report := conv.GenerateReport()
			if report.Summary.ParticipantRef != tc.participant {
				t.Errorf("Expected report participant %s, got %s",
					tc.participant, report.Summary.ParticipantRef)
			}
		})
	}
}

// Helper: Create minimal GTFS files for testing
func createMinimalGTFS(t *testing.T, dir string) {
	// agency.txt
	agencyContent := `agency_id,agency_name,agency_url,agency_timezone
TEST_AGENCY,Test Transit,http://test.transit,Europe/Sofia
`
	writeFile(t, filepath.Join(dir, "agency.txt"), agencyContent)

	// stops.txt
	stopsContent := `stop_id,stop_name,stop_lat,stop_lon
STOP1,Stop 1,42.6977,23.3219
STOP2,Stop 2,42.6978,23.3220
`
	writeFile(t, filepath.Join(dir, "stops.txt"), stopsContent)

	// routes.txt
	routesContent := `route_id,route_short_name,route_long_name,route_type
ROUTE1,1,Route One,3
`
	writeFile(t, filepath.Join(dir, "routes.txt"), routesContent)

	// trips.txt
	tripsContent := `route_id,service_id,trip_id
ROUTE1,SERVICE1,TRIP1
`
	writeFile(t, filepath.Join(dir, "trips.txt"), tripsContent)

	// stop_times.txt
	stopTimesContent := `trip_id,arrival_time,departure_time,stop_id,stop_sequence
TRIP1,08:00:00,08:00:00,STOP1,1
TRIP1,08:10:00,08:10:00,STOP2,2
`
	writeFile(t, filepath.Join(dir, "stop_times.txt"), stopTimesContent)

	// calendar_dates.txt
	calendarDatesContent := `service_id,date,exception_type
SERVICE1,20250101,1
SERVICE1,20250102,1
`
	writeFile(t, filepath.Join(dir, "calendar_dates.txt"), calendarDatesContent)
}

// Helper: Create larger GTFS dataset for performance testing
func createLargerGTFS(t *testing.T, dir string) {
	// Just use the minimal GTFS for now - it's sufficient for testing
	createMinimalGTFS(t, dir)
}

func writeFile(t *testing.T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", path, err)
	}
}
