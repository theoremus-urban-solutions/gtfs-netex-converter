package main

import (
	"archive/zip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/converter"
	"github.com/theoremus-urban-solutions/gtfs-netex-converter/stream"
)

func main() {
	// Command line flags
	var (
		input          = flag.String("input", "", "GTFS zip or directory of GTFS files")
		outputFile     = flag.String("output", "output.xml", "Output NeTEx XML file; a .zip name zips it")
		pubTime        = flag.String("publication-time", "", "PublicationTimestamp as 2006-01-02T15:04:05 (default now)")
		reportFile     = flag.String("report", "", "Output JSON report file (optional)")
		reportFormat   = flag.String("report-format", "text", "Report format: text, json, or both")
		participantRef = flag.String("participant", "BG", "Codespace and participant reference (e.g., BG for Bulgaria)")
		timezone       = flag.String("timezone", "Europe/Sofia", "Default timezone")
		language       = flag.String("language", "bg", "Default language")
		verbose        = flag.Bool("verbose", false, "Enable verbose output")
	)
	// Flags of the old tree converter, kept so existing scripts still run.
	flag.String("location-system", "EPSG:4326", "Ignored since v1.1.0: output is always EPSG:4326")
	flag.Bool("fare-frame", false, "Ignored since v1.1.0: fares are outside EPIP")
	flag.Bool("general-frame", false, "Ignored since v1.1.0")
	flag.Bool("streaming", false, "Ignored since v1.1.0: conversion always streams")
	flag.Int("batch-size", 10000, "Ignored since v1.1.0")

	flag.Parse()

	if *input == "" {
		fmt.Fprintf(os.Stderr, "Error: Input is required. Use -input flag.\n\n")
		flag.Usage()
		os.Exit(1)
	}
	st, err := os.Stat(*input)
	if err != nil {
		log.Fatalf("Error: Input does not exist: %s", *input)
	}

	// Create output directory if needed
	outputDir := filepath.Dir(*outputFile)
	if outputDir != "." && outputDir != "" {
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			log.Fatalf("Error: Failed to create output directory: %v", err)
		}
	}

	if *verbose {
		fmt.Println("GTFS to NeTEx Converter (EPIP)")
		fmt.Printf("  Input:       %s\n", *input)
		fmt.Printf("  Output:      %s\n", *outputFile)
		fmt.Printf("  Participant: %s\n", *participantRef)
		fmt.Printf("  Timezone:    %s\n", *timezone)
		fmt.Printf("  Language:    %s\n", *language)
		fmt.Println()
	}

	cfg := stream.Config{Codespace: *participantRef, TimeZone: *timezone, Language: *language}
	if *pubTime != "" {
		t, err := time.Parse("2006-01-02T15:04:05", *pubTime)
		if err != nil {
			log.Fatalf("Error: bad -publication-time: %v", err)
		}
		cfg.PublicationTime = t
	}
	report, err := convert(*input, st.IsDir(), *outputFile, cfg)
	if err != nil {
		log.Fatalf("Conversion failed: %v", err)
	}
	report.Summary.InputDirectory = *input
	report.Summary.OutputFile = *outputFile
	writeReport(report, *reportFile, *reportFormat, *verbose)
}

// convert runs the streaming EPIP converter; a .zip output gets one XML entry named after it.
func convert(input string, isDir bool, output string, cfg stream.Config) (*converter.ConversionReport, error) {
	f, err := os.Create(output) //nolint:gosec // output path is the user's choice
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	var dst io.Writer = f
	var zw *zip.Writer
	if strings.HasSuffix(strings.ToLower(output), ".zip") {
		zw = zip.NewWriter(f)
		name := strings.TrimSuffix(filepath.Base(output), filepath.Ext(output)) + ".xml"
		if dst, err = zw.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate, Modified: time.Now()}); err != nil {
			return nil, err
		}
	}
	var report *converter.ConversionReport
	if isDir {
		report, err = stream.ConvertFS(context.Background(), os.DirFS(input), dst, cfg)
	} else {
		report, err = stream.ConvertZipFile(context.Background(), input, dst, cfg)
	}
	if err != nil {
		return nil, err
	}
	if zw != nil {
		if err := zw.Close(); err != nil {
			return nil, err
		}
	}
	return report, f.Close()
}

func writeReport(report *converter.ConversionReport, reportFile, reportFormat string, verbose bool) {
	// Output report based on format
	shouldOutputText := reportFormat == converter.ReportFormatText || reportFormat == converter.ReportFormatBoth || (reportFormat != converter.ReportFormatJSON && reportFile == "")
	shouldOutputJSON := reportFormat == converter.ReportFormatJSON || reportFormat == converter.ReportFormatBoth || reportFile != ""

	// Text output (console or verbose)
	if shouldOutputText && verbose {
		printTextReport(report)
	} else if shouldOutputText && !verbose {
		fmt.Printf("Conversion completed: %s\n", report.Summary.OutputFile)
		fmt.Printf("  Processed %d GTFS entities → %d NeTEx entities in %.2fs\n",
			report.InputStatistics.TotalRecords,
			report.OutputStatistics.TotalEntities,
			report.Summary.DurationSeconds)
	}

	// JSON output to file
	if shouldOutputJSON && reportFile != "" {
		jsonData, err := report.ToJSON()
		if err != nil {
			log.Fatalf("Failed to generate JSON report: %v", err)
		}

		if err := os.WriteFile(reportFile, jsonData, 0600); err != nil {
			log.Fatalf("Failed to write report file: %v", err)
		}

		if verbose {
			fmt.Printf("\n📊 JSON report saved to: %s\n", reportFile)
		}
	}

	// JSON output to console
	if shouldOutputJSON && reportFile == "" && reportFormat == "json" {
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
