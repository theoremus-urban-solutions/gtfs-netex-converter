package converter

import (
	"fmt"
	"time"

	"gtfs-netex-converter/gtfs"
	"gtfs-netex-converter/netex"
)

// CalendarConverter handles conversion of GTFS calendar data to NeTEx calendar entities
type CalendarConverter struct {
	idMapper *IDMapper
}

// NewCalendarConverter creates a new calendar converter
func NewCalendarConverter(idMapper *IDMapper) *CalendarConverter {
	return &CalendarConverter{
		idMapper: idMapper,
	}
}

// ConvertCalendar converts GTFS calendar data to NeTEx calendar entities
// Since Sofia lacks calendar.txt, we infer patterns from calendar_dates.txt and trips
func (cc *CalendarConverter) ConvertCalendar(calendarDates []gtfs.CalendarDate, trips []gtfs.Trip) (*CalendarConversionResult, error) {
	fmt.Println("Converting calendar data...")

	result := &CalendarConversionResult{
		DayTypes:           []netex.DayType{},
		OperatingPeriods:   []netex.UicOperatingPeriod{},
		DayTypeAssignments: []netex.DayTypeAssignment{},
	}

	// Since Sofia lacks calendar.txt, we need to infer service patterns
	servicePatterns := cc.inferServicePatterns(trips, calendarDates)

	// Generate DayTypes and OperatingPeriods for each service pattern
	for serviceID, pattern := range servicePatterns {
		// Validate the pattern has valid dates
		if pattern.StartDate == "" || pattern.EndDate == "" ||
			pattern.StartDate == DateMax || pattern.EndDate == DateMin {
			fmt.Printf("Warning: Invalid pattern for service %s, using default\n", serviceID)
			pattern = cc.createDefaultPattern(serviceID)
		}

		dayType := cc.createDayType(serviceID)
		result.DayTypes = append(result.DayTypes, dayType)

		operatingPeriod := cc.createOperatingPeriod(serviceID, pattern)
		result.OperatingPeriods = append(result.OperatingPeriods, operatingPeriod)

		dayTypeAssignment := cc.createDayTypeAssignment(serviceID, dayType, operatingPeriod)
		result.DayTypeAssignments = append(result.DayTypeAssignments, dayTypeAssignment)
	}

	// If no patterns were found, create a default one
	if len(result.DayTypes) == 0 {
		fmt.Println("No calendar patterns found, creating default pattern...")
		defaultDayType := netex.DayType{
			ID:      "BG::DayType:DAILY::",
			Version: "1",
		}
		result.DayTypes = append(result.DayTypes, defaultDayType)

		// Format current date properly
		now := time.Now()
		fromDate := now.Format("2006-01-02")
		toDate := now.AddDate(1, 0, 0).Format("2006-01-02")

		defaultOperatingPeriod := netex.UicOperatingPeriod{
			ID:           "BG::OperatingPeriod:DAILY::",
			Version:      "1",
			FromDate:     fromDate + "T00:00:00",
			ToDate:       toDate + "T00:00:00",
			ValidDayBits: "1111111", // Daily service
		}
		result.OperatingPeriods = append(result.OperatingPeriods, defaultOperatingPeriod)

		defaultDayTypeAssignment := netex.DayTypeAssignment{
			ID:      "BG::DayTypeAssignment:DAILY::",
			Order:   "0",
			Version: "1",
			OperatingPeriodRef: netex.OperatingPeriodRef{
				Ref:     defaultOperatingPeriod.ID,
				Version: defaultOperatingPeriod.Version,
			},
			DayTypeRef: netex.DayTypeRef{
				Ref:     defaultDayType.ID,
				Version: defaultDayType.Version,
			},
			IsAvailable: true,
		}
		result.DayTypeAssignments = append(result.DayTypeAssignments, defaultDayTypeAssignment)
	}

	fmt.Printf("Converted calendar data to %d DayTypes, %d OperatingPeriods, %d DayTypeAssignments\n",
		len(result.DayTypes), len(result.OperatingPeriods), len(result.DayTypeAssignments))

	return result, nil
}

// CalendarConversionResult holds the result of calendar conversion
type CalendarConversionResult struct {
	DayTypes           []netex.DayType
	OperatingPeriods   []netex.UicOperatingPeriod
	DayTypeAssignments []netex.DayTypeAssignment
}

// ServicePattern represents inferred service pattern
type ServicePattern struct {
	ServiceID    string
	StartDate    string
	EndDate      string
	ValidDayBits string         // UIC format: 1111100 for Mon-Fri
	Exceptions   map[string]int // date -> exception_type
}

// inferServicePatterns infers service patterns from trips and calendar_dates
func (cc *CalendarConverter) inferServicePatterns(trips []gtfs.Trip, calendarDates []gtfs.CalendarDate) map[string]ServicePattern {
	patterns := make(map[string]ServicePattern)

	// Group calendar dates by service_id
	datesByService := make(map[string][]gtfs.CalendarDate)
	for _, date := range calendarDates {
		datesByService[date.ServiceID] = append(datesByService[date.ServiceID], date)
	}

	// For each service, infer pattern
	for serviceID, dates := range datesByService {
		pattern := cc.inferPatternForService(serviceID, dates)
		patterns[serviceID] = pattern
	}

	// For services without calendar_dates, create default pattern
	for _, trip := range trips {
		if _, exists := patterns[trip.ServiceID]; !exists {
			pattern := cc.createDefaultPattern(trip.ServiceID)
			patterns[trip.ServiceID] = pattern
		}
	}

	return patterns
}

// inferPatternForService infers pattern for a specific service
func (cc *CalendarConverter) inferPatternForService(serviceID string, dates []gtfs.CalendarDate) ServicePattern {
	pattern := ServicePattern{
		ServiceID:  serviceID,
		Exceptions: make(map[string]int),
	}

	// If no calendar dates, use default pattern
	if len(dates) == 0 {
		return cc.createDefaultPattern(serviceID)
	}

	// Find date range
	minDate := DateMax
	maxDate := DateMin

	for _, date := range dates {
		if date.Date < minDate {
			minDate = date.Date
		}
		if date.Date > maxDate {
			maxDate = date.Date
		}
		pattern.Exceptions[date.Date] = date.ExceptionType
	}

	// Validate that we found valid dates
	if minDate == DateMax || maxDate == DateMin {
		// No valid dates found, use default pattern
		return cc.createDefaultPattern(serviceID)
	}

	pattern.StartDate = minDate
	pattern.EndDate = maxDate

	// For Sofia urban transit, assume daily service (1111111)
	// This is a simplification - in reality you'd analyze trip patterns
	pattern.ValidDayBits = "1111111" // Daily service

	return pattern
}

// createDefaultPattern creates a default pattern for services without calendar_dates
func (cc *CalendarConverter) createDefaultPattern(serviceID string) ServicePattern {
	// For Sofia urban transit, assume daily service for the current year
	now := time.Now()
	startDate := fmt.Sprintf("%d-01-01", now.Year())
	endDate := fmt.Sprintf("%d-12-31", now.Year())

	return ServicePattern{
		ServiceID:    serviceID,
		StartDate:    startDate,
		EndDate:      endDate,
		ValidDayBits: "1111111", // Daily service
		Exceptions:   make(map[string]int),
	}
}

// createDayType creates a DayType for a service
func (cc *CalendarConverter) createDayType(serviceID string) netex.DayType {
	return netex.DayType{
		ID:      fmt.Sprintf("BG::DayType:day_type_%s::", serviceID),
		Version: "1",
		// Note: NeTEx DayType doesn't have Properties in our struct
		// Could be extended to include day-of-week properties
	}
}

// createOperatingPeriod creates an OperatingPeriod for a service
func (cc *CalendarConverter) createOperatingPeriod(serviceID string, pattern ServicePattern) netex.UicOperatingPeriod {
	// Validate dates before creating the operating period
	fromDate := pattern.StartDate
	toDate := pattern.EndDate

	// If dates are invalid, use current year
	if fromDate == "" || fromDate == DateMax || toDate == "" || toDate == DateMin {
		now := time.Now()
		fromDate = fmt.Sprintf("%d-01-01", now.Year())
		toDate = fmt.Sprintf("%d-12-31", now.Year())
	}

	// Convert GTFS date format (YYYYMMDD) to ISO format (YYYY-MM-DD) if needed
	fromDateFormatted := cc.formatDateForNeTEx(fromDate)
	toDateFormatted := cc.formatDateForNeTEx(toDate)

	return netex.UicOperatingPeriod{
		ID:           fmt.Sprintf("BG::OperatingPeriod:operating_period_%s::", serviceID),
		Version:      "1",
		FromDate:     fromDateFormatted + "T00:00:00",
		ToDate:       toDateFormatted + "T00:00:00",
		ValidDayBits: pattern.ValidDayBits,
	}
}

// formatDateForNeTEx converts various date formats to ISO format (YYYY-MM-DD)
func (cc *CalendarConverter) formatDateForNeTEx(dateStr string) string {
	// If already in ISO format (YYYY-MM-DD), return as is
	if len(dateStr) == 10 && dateStr[4] == '-' && dateStr[7] == '-' {
		return dateStr
	}

	// If in GTFS format (YYYYMMDD), convert to ISO format
	if len(dateStr) == 8 {
		return dateStr[:4] + "-" + dateStr[4:6] + "-" + dateStr[6:8]
	}

	// If it's a valid date, try to parse and format it
	if len(dateStr) >= 8 {
		// Try to parse as YYYYMMDD
		if len(dateStr) == 8 {
			year := dateStr[:4]
			month := dateStr[4:6]
			day := dateStr[6:8]
			return year + "-" + month + "-" + day
		}
	}

	// If all else fails, return the original string
	return dateStr
}

// createDayTypeAssignment creates a DayTypeAssignment linking DayType and OperatingPeriod
func (cc *CalendarConverter) createDayTypeAssignment(serviceID string, dayType netex.DayType, operatingPeriod netex.UicOperatingPeriod) netex.DayTypeAssignment {
	return netex.DayTypeAssignment{
		ID:      fmt.Sprintf("day_type_assignment_%s", serviceID),
		Order:   "0",
		Version: "1",
		OperatingPeriodRef: netex.OperatingPeriodRef{
			Ref:     operatingPeriod.ID,
			Version: operatingPeriod.Version,
		},
		DayTypeRef: netex.DayTypeRef{
			Ref:     dayType.ID,
			Version: dayType.Version,
		},
		IsAvailable: true, // Default to available
	}
}

// GetDayTypeRef returns a DayTypeRef for a service
func (cc *CalendarConverter) GetDayTypeRef(serviceID string) netex.DayTypeRef {
	return netex.DayTypeRef{
		Ref:     fmt.Sprintf("BG::DayType:day_type_%s::", serviceID),
		Version: "1",
	}
}
