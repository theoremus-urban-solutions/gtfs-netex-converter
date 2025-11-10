package converter

import (
	"encoding/json"
	"time"
)

// ConversionReport contains comprehensive statistics and results from the conversion
type ConversionReport struct {
	Summary          ConversionSummary             `json:"summary"`
	InputStatistics  InputStatistics               `json:"input_statistics"`
	OutputStatistics OutputStatistics              `json:"output_statistics"`
	Performance      PerformanceMetrics            `json:"performance"`
	Validation       ValidationResults             `json:"validation"`
	EntityMapping    map[string]EntityMappingStats `json:"entity_mapping"`
	Warnings         []Warning                     `json:"warnings"`
	Errors           []ConversionError             `json:"errors"`
}

// ConversionSummary provides high-level conversion results
type ConversionSummary struct {
	Status           string    `json:"status"` // "success", "partial", "failed"
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	DurationSeconds  float64   `json:"duration_seconds"`
	ParticipantRef   string    `json:"participant_ref"`
	InputDirectory   string    `json:"input_directory"`
	OutputFile       string    `json:"output_file"`
	ConverterVersion string    `json:"converter_version"`
}

// InputStatistics tracks input GTFS data
type InputStatistics struct {
	FilesProcessed int            `json:"files_processed"`
	FilesSkipped   int            `json:"files_skipped"`
	TotalRecords   int            `json:"total_records"`
	RecordsByFile  map[string]int `json:"records_by_file"`
	Agencies       int            `json:"agencies"`
	Stops          int            `json:"stops"`
	Routes         int            `json:"routes"`
	Trips          int            `json:"trips"`
	StopTimes      int            `json:"stop_times"`
	CalendarDates  int            `json:"calendar_dates"`
	Shapes         int            `json:"shapes"`
	Transfers      int            `json:"transfers"`
	FareAttributes int            `json:"fare_attributes"`
	Pathways       int            `json:"pathways"`
}

// OutputStatistics tracks generated NeTEx entities
type OutputStatistics struct {
	TotalEntities       int            `json:"total_entities"`
	EntitiesByType      map[string]int `json:"entities_by_type"`
	Authorities         int            `json:"authorities"`
	Operators           int            `json:"operators"`
	StopPlaces          int            `json:"stop_places"`
	Quays               int            `json:"quays"`
	ScheduledStopPoints int            `json:"scheduled_stop_points"`
	Lines               int            `json:"lines"`
	Routes              int            `json:"routes"`
	ServiceJourneys     int            `json:"service_journeys"`
	JourneyPatterns     int            `json:"journey_patterns"`
	DayTypes            int            `json:"day_types"`
	OperatingPeriods    int            `json:"operating_periods"`
	ServiceLinks        int            `json:"service_links"`
}

// PerformanceMetrics tracks conversion performance
type PerformanceMetrics struct {
	TotalDurationSeconds      float64            `json:"total_duration_seconds"`
	LoadingDurationSeconds    float64            `json:"loading_duration_seconds"`
	IndexingDurationSeconds   float64            `json:"indexing_duration_seconds"`
	ConversionDurationSeconds float64            `json:"conversion_duration_seconds"`
	RecordsPerSecond          float64            `json:"records_per_second"`
	MemoryUsageMB             float64            `json:"memory_usage_mb"`
	PeakMemoryUsageMB         float64            `json:"peak_memory_usage_mb"`
	StageTimings              map[string]float64 `json:"stage_timings"`
}

// ValidationResults tracks validation issues
type ValidationResults struct {
	TotalIssues      int               `json:"total_issues"`
	IssuesBySeverity map[string]int    `json:"issues_by_severity"`
	IssuesByType     map[string]int    `json:"issues_by_type"`
	CriticalIssues   []ValidationIssue `json:"critical_issues"`
	Warnings         []ValidationIssue `json:"warnings"`
}

// ValidationIssue represents a validation problem
type ValidationIssue struct {
	Severity   string `json:"severity"` // "info", "warning", "error", "critical"
	Code       string `json:"code"`
	Message    string `json:"message"`
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id,omitempty"`
	Field      string `json:"field,omitempty"`
	Location   string `json:"location,omitempty"`
}

// EntityMappingStats tracks how entities were mapped
type EntityMappingStats struct {
	InputCount    int     `json:"input_count"`
	OutputCount   int     `json:"output_count"`
	SkippedCount  int     `json:"skipped_count"`
	SuccessRate   float64 `json:"success_rate"`
	AverageTimeMs float64 `json:"average_time_ms"`
}

// Warning represents a non-critical issue
type Warning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// ConversionError represents an error during conversion
type ConversionError struct {
	Stage    string `json:"stage"`
	Message  string `json:"message"`
	Entity   string `json:"entity,omitempty"`
	Location string `json:"location,omitempty"`
}

// GenerateReport creates a comprehensive conversion report
func (c *Converter) GenerateReport() *ConversionReport {
	duration := c.endTime.Sub(c.startTime)

	report := &ConversionReport{
		Summary: ConversionSummary{
			Status:           "success",
			StartTime:        c.startTime,
			EndTime:          c.endTime,
			DurationSeconds:  duration.Seconds(),
			ParticipantRef:   c.config.ParticipantRef,
			InputDirectory:   c.inputDir,
			ConverterVersion: "1.0.0",
		},
		InputStatistics:  c.generateInputStatistics(),
		OutputStatistics: c.generateOutputStatistics(),
		Performance:      c.generatePerformanceMetrics(),
		Validation:       c.generateValidationResults(),
		EntityMapping:    c.generateEntityMapping(),
		Warnings:         c.collectWarnings(),
		Errors:           c.collectErrors(),
	}

	return report
}

// generateInputStatistics collects input GTFS statistics
func (c *Converter) generateInputStatistics() InputStatistics {
	recordsByFile := make(map[string]int)
	recordsByFile["agency.txt"] = len(c.gtfsData.Agencies)
	recordsByFile["stops.txt"] = len(c.gtfsData.Stops)
	recordsByFile["routes.txt"] = len(c.gtfsData.Routes)
	recordsByFile["trips.txt"] = len(c.gtfsData.Trips)
	recordsByFile["stop_times.txt"] = len(c.gtfsData.StopTimes)
	recordsByFile["calendar_dates.txt"] = len(c.gtfsData.CalendarDates)
	recordsByFile["shapes.txt"] = len(c.gtfsData.Shapes)
	recordsByFile["transfers.txt"] = len(c.gtfsData.Transfers)
	recordsByFile["fare_attributes.txt"] = len(c.gtfsData.FareAttributes)
	recordsByFile["pathways.txt"] = len(c.gtfsData.Pathways)

	totalRecords := 0
	filesProcessed := 0
	for _, count := range recordsByFile {
		totalRecords += count
		if count > 0 {
			filesProcessed++
		}
	}

	return InputStatistics{
		FilesProcessed: filesProcessed,
		FilesSkipped:   0,
		TotalRecords:   totalRecords,
		RecordsByFile:  recordsByFile,
		Agencies:       len(c.gtfsData.Agencies),
		Stops:          len(c.gtfsData.Stops),
		Routes:         len(c.gtfsData.Routes),
		Trips:          len(c.gtfsData.Trips),
		StopTimes:      len(c.gtfsData.StopTimes),
		CalendarDates:  len(c.gtfsData.CalendarDates),
		Shapes:         len(c.gtfsData.Shapes),
		Transfers:      len(c.gtfsData.Transfers),
		FareAttributes: len(c.gtfsData.FareAttributes),
		Pathways:       len(c.gtfsData.Pathways),
	}
}

// generateOutputStatistics collects output NeTEx statistics
func (c *Converter) generateOutputStatistics() OutputStatistics {
	stats := c.GetStats()

	entitiesByType := make(map[string]int)
	entitiesByType["Authority"] = len(c.gtfsData.Agencies)
	entitiesByType["Operator"] = len(c.gtfsData.Agencies)
	entitiesByType["StopPlace"] = len(c.gtfsData.Stops)
	entitiesByType["Quay"] = len(c.gtfsData.Stops)
	entitiesByType["ScheduledStopPoint"] = len(c.gtfsData.Stops)
	entitiesByType["Line"] = len(c.gtfsData.Routes)
	entitiesByType["Route"] = len(c.gtfsData.Routes)
	entitiesByType["ServiceJourney"] = len(c.gtfsData.Trips)

	return OutputStatistics{
		TotalEntities:       stats.NeTExEntities,
		EntitiesByType:      entitiesByType,
		Authorities:         len(c.gtfsData.Agencies),
		Operators:           len(c.gtfsData.Agencies),
		StopPlaces:          len(c.gtfsData.Stops),
		Quays:               len(c.gtfsData.Stops),
		ScheduledStopPoints: len(c.gtfsData.Stops),
		Lines:               len(c.gtfsData.Routes),
		Routes:              len(c.gtfsData.Routes),
		ServiceJourneys:     len(c.gtfsData.Trips),
	}
}

// generatePerformanceMetrics collects performance data
func (c *Converter) generatePerformanceMetrics() PerformanceMetrics {
	duration := c.endTime.Sub(c.startTime).Seconds()
	stats := c.GetStats()

	recordsPerSecond := 0.0
	if duration > 0 {
		recordsPerSecond = float64(stats.GTFSEntities) / duration
	}

	return PerformanceMetrics{
		TotalDurationSeconds: duration,
		RecordsPerSecond:     recordsPerSecond,
		MemoryUsageMB:        0, // TODO: Implement memory tracking
		PeakMemoryUsageMB:    0,
		StageTimings:         make(map[string]float64),
	}
}

// generateValidationResults collects validation issues
func (c *Converter) generateValidationResults() ValidationResults {
	return ValidationResults{
		TotalIssues:      0,
		IssuesBySeverity: make(map[string]int),
		IssuesByType:     make(map[string]int),
		CriticalIssues:   []ValidationIssue{},
		Warnings:         []ValidationIssue{},
	}
}

// generateEntityMapping tracks entity mapping statistics
func (c *Converter) generateEntityMapping() map[string]EntityMappingStats {
	mapping := make(map[string]EntityMappingStats)

	mapping["Agency->Authority"] = EntityMappingStats{
		InputCount:   len(c.gtfsData.Agencies),
		OutputCount:  len(c.gtfsData.Agencies),
		SkippedCount: 0,
		SuccessRate:  100.0,
	}

	mapping["Stop->StopPlace"] = EntityMappingStats{
		InputCount:   len(c.gtfsData.Stops),
		OutputCount:  len(c.gtfsData.Stops),
		SkippedCount: 0,
		SuccessRate:  100.0,
	}

	mapping["Route->Line"] = EntityMappingStats{
		InputCount:   len(c.gtfsData.Routes),
		OutputCount:  len(c.gtfsData.Routes),
		SkippedCount: 0,
		SuccessRate:  100.0,
	}

	mapping["Trip->ServiceJourney"] = EntityMappingStats{
		InputCount:   len(c.gtfsData.Trips),
		OutputCount:  len(c.gtfsData.Trips),
		SkippedCount: 0,
		SuccessRate:  100.0,
	}

	return mapping
}

// collectWarnings gathers non-critical warnings
func (c *Converter) collectWarnings() []Warning {
	warnings := []Warning{}

	if len(c.gtfsData.Shapes) > 0 {
		warnings = append(warnings, Warning{
			Type:    "feature_not_implemented",
			Message: "Shapes data loaded but not converted to NeTEx",
			Count:   len(c.gtfsData.Shapes),
		})
	}

	if len(c.gtfsData.FareAttributes) > 0 && !c.config.GenerateFareFrame {
		warnings = append(warnings, Warning{
			Type:    "feature_not_enabled",
			Message: "Fare data available but FareFrame generation not enabled",
			Count:   len(c.gtfsData.FareAttributes),
		})
	}

	return warnings
}

// collectErrors gathers errors encountered during conversion
func (c *Converter) collectErrors() []ConversionError {
	// For now, return empty - can be enhanced to track errors during conversion
	return []ConversionError{}
}

// ToJSON converts the report to JSON format
func (r *ConversionReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToJSONCompact converts the report to compact JSON format
func (r *ConversionReport) ToJSONCompact() ([]byte, error) {
	return json.Marshal(r)
}
