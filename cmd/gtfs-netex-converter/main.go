package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gtfs-netex-converter/converter"
)

func main() {
	// Command line flags
	var (
		inputDir             = flag.String("input", "", "Input directory containing GTFS files")
		outputFile           = flag.String("output", "output.xml", "Output NeTEx XML file")
		reportFile           = flag.String("report", "", "Output JSON report file (optional)")
		reportFormat         = flag.String("report-format", "text", "Report format: text, json, or both")
		participantRef       = flag.String("participant", "BG", "Participant reference (e.g., BG for Bulgaria)")
		timezone             = flag.String("timezone", "Europe/Sofia", "Default timezone")
		language             = flag.String("language", "bg", "Default language")
		locationSystem       = flag.String("location-system", "EPSG:4326", "Location coordinate system")
		generateFareFrame    = flag.Bool("fare-frame", false, "Generate FareFrame (if fare data available)")
		generateGeneralFrame = flag.Bool("general-frame", false, "Generate GeneralFrame (if infrastructure data available)")
		verbose              = flag.Bool("verbose", false, "Enable verbose output")
	)

	flag.Parse()

	// Validate required arguments
	if *inputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: Input directory is required. Use -input flag.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if input directory exists
	if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
		log.Fatalf("Error: Input directory does not exist: %s", *inputDir)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(*outputFile)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatalf("Error: Failed to create output directory: %v", err)
		}
	}

	// Print configuration if verbose
	if *verbose {
		fmt.Println("GTFS to NeTEx Converter")
		fmt.Printf("  Input:       %s\n", *inputDir)
		fmt.Printf("  Output:      %s\n", *outputFile)
		fmt.Printf("  Participant: %s\n", *participantRef)
		fmt.Printf("  Timezone:    %s\n", *timezone)
		fmt.Printf("  Language:    %s\n", *language)
		fmt.Println()
	}

	// Create converter configuration
	config := &converter.Config{
		ParticipantRef:       *participantRef,
		DefaultTimezone:      *timezone,
		DefaultLanguage:      *language,
		LocationSystem:       *locationSystem,
		GenerateFareFrame:    *generateFareFrame,
		GenerateGeneralFrame: *generateGeneralFrame,
	}

	// Create and run converter
	conv := converter.NewConverter(config)
	netexData, err := conv.Convert(*inputDir)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}

	// Generate XML output
	xmlData, err := xml.MarshalIndent(netexData, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal XML: %v", err)
	}

	// Write output file
	xmlOutput := xml.Header + string(xmlData)
	if err := os.WriteFile(*outputFile, []byte(xmlOutput), 0644); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	// Generate comprehensive report
	report := conv.GenerateReport()
	report.Summary.OutputFile = *outputFile

	// Output report based on format
	shouldOutputText := *reportFormat == "text" || *reportFormat == "both" || (*reportFormat != "json" && *reportFile == "")
	shouldOutputJSON := *reportFormat == "json" || *reportFormat == "both" || *reportFile != ""

	// Text output (console or verbose)
	if shouldOutputText && *verbose {
		printTextReport(report)
	} else if shouldOutputText && !*verbose {
		fmt.Printf("Conversion completed: %s\n", *outputFile)
		fmt.Printf("  Processed %d GTFS entities → %d NeTEx entities in %.2fs\n",
			report.InputStatistics.TotalRecords,
			report.OutputStatistics.TotalEntities,
			report.Summary.DurationSeconds)
	}

	// JSON output to file
	if shouldOutputJSON && *reportFile != "" {
		jsonData, err := report.ToJSON()
		if err != nil {
			log.Fatalf("Failed to generate JSON report: %v", err)
		}

		if err := os.WriteFile(*reportFile, jsonData, 0644); err != nil {
			log.Fatalf("Failed to write report file: %v", err)
		}

		if *verbose {
			fmt.Printf("\n📊 JSON report saved to: %s\n", *reportFile)
		}
	}

	// JSON output to console
	if shouldOutputJSON && *reportFile == "" && *reportFormat == "json" {
		jsonData, err := report.ToJSON()
		if err != nil {
			log.Fatalf("Failed to generate JSON report: %v", err)
		}
		fmt.Println(string(jsonData))
	}
}

// printTextReport prints a detailed text report to console
func printTextReport(report *converter.ConversionReport) {
	fmt.Printf("\n✅ Conversion completed successfully!\n")
	fmt.Printf("   Output: %s\n", report.Summary.OutputFile)

	fmt.Printf("\n📥 Input GTFS Statistics:\n")
	fmt.Printf("  Files processed:  %d\n", report.InputStatistics.FilesProcessed)
	fmt.Printf("  Total records:    %d\n", report.InputStatistics.TotalRecords)
	fmt.Printf("  Agencies:         %d\n", report.InputStatistics.Agencies)
	fmt.Printf("  Stops:            %d\n", report.InputStatistics.Stops)
	fmt.Printf("  Routes:           %d\n", report.InputStatistics.Routes)
	fmt.Printf("  Trips:            %d\n", report.InputStatistics.Trips)
	fmt.Printf("  Stop times:       %d\n", report.InputStatistics.StopTimes)

	fmt.Printf("\n📤 Output NeTEx Statistics:\n")
	fmt.Printf("  Total entities:   %d\n", report.OutputStatistics.TotalEntities)
	fmt.Printf("  Authorities:      %d\n", report.OutputStatistics.Authorities)
	fmt.Printf("  Stop places:      %d\n", report.OutputStatistics.StopPlaces)
	fmt.Printf("  Lines:            %d\n", report.OutputStatistics.Lines)
	fmt.Printf("  Service journeys: %d\n", report.OutputStatistics.ServiceJourneys)

	fmt.Printf("\n⚡ Performance:\n")
	fmt.Printf("  Duration:         %.2fs\n", report.Performance.TotalDurationSeconds)
	fmt.Printf("  Processing rate:  %.0f entities/s\n", report.Performance.RecordsPerSecond)

	if len(report.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(report.Warnings))
		for _, warning := range report.Warnings {
			fmt.Printf("  - %s: %s (%d)\n", warning.Type, warning.Message, warning.Count)
		}
	}

	if len(report.Errors) > 0 {
		fmt.Printf("\n❌ Errors (%d):\n", len(report.Errors))
		for _, err := range report.Errors {
			fmt.Printf("  - [%s] %s\n", err.Stage, err.Message)
		}
	}
}
