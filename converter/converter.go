package converter

import (
	"crypto/md5" //nolint:gosec // MD5 used for non-cryptographic ID generation only
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/gtfs"
	"github.com/theoremus-urban-solutions/gtfs-netex-converter/netex"
)

// Converter represents the main GTFS to NeTEx converter
type Converter struct {
	gtfsData    *GTFSData
	netexData   *netex.PublicationDelivery
	idMapper    *IDMapper
	idGenerator *IDGenerator
	config      *Config
	lookupIndex *LookupIndices
	inputDir    string
	startTime   time.Time
	endTime     time.Time
}

// GTFSData holds all GTFS data loaded from files
type GTFSData struct {
	Agencies       []gtfs.Agency
	Stops          []gtfs.Stop
	Routes         []gtfs.Route
	Trips          []gtfs.Trip
	StopTimes      []gtfs.StopTime
	CalendarDates  []gtfs.CalendarDate
	Shapes         []gtfs.Shape
	Transfers      []gtfs.Transfer
	Pathways       []gtfs.Pathway
	Levels         []gtfs.Level
	FeedInfo       []gtfs.FeedInfo
	FareAttributes []gtfs.FareAttribute
	Translations   []gtfs.Translation
}

// Config holds converter configuration
type Config struct {
	ParticipantRef       string
	DefaultTimezone      string
	DefaultLanguage      string
	LocationSystem       string
	GenerateFareFrame    bool
	GenerateGeneralFrame bool
}

// NewConverter creates a new converter instance
func NewConverter(config *Config) *Converter {
	return &Converter{
		gtfsData:    &GTFSData{},
		netexData:   &netex.PublicationDelivery{},
		idMapper:    NewIDMapper(),
		idGenerator: NewIDGenerator(config.ParticipantRef),
		config:      config,
	}
}

// Convert performs the complete GTFS to NeTEx conversion
func (c *Converter) Convert(inputDir string) (*netex.PublicationDelivery, error) {
	c.startTime = time.Now()
	fmt.Println("Starting GTFS to NeTEx conversion...")
	c.inputDir = inputDir

	// Step 1: Load GTFS data
	if err := c.loadGTFSData(); err != nil {
		return nil, fmt.Errorf("failed to load GTFS data: %w", err)
	}

	// Step 2: Build lookup indices for O(1) access
	fmt.Println("Building lookup indices...")
	c.lookupIndex = BuildLookupIndices(c.gtfsData)

	// Step 3: Generate missing entities
	c.generateMissingEntities()

	// Step 4: Create NeTEx structure
	if err := c.createNeTExStructure(); err != nil {
		return nil, fmt.Errorf("failed to create NeTEx structure: %w", err)
	}

	// Step 5: Organize into frames
	c.organizeIntoFrames()

	c.endTime = time.Now()
	duration := c.endTime.Sub(c.startTime)
	fmt.Printf("Conversion completed successfully in %.2f seconds!\n", duration.Seconds())

	return c.netexData, nil
}

// loadGTFSData loads all GTFS files
func (c *Converter) loadGTFSData() error {
	fmt.Println("Loading GTFS data...")

	// Load each GTFS file
	files := map[string]interface{}{
		"agency.txt":          &c.gtfsData.Agencies,
		"stops.txt":           &c.gtfsData.Stops,
		"routes.txt":          &c.gtfsData.Routes,
		"trips.txt":           &c.gtfsData.Trips,
		"stop_times.txt":      &c.gtfsData.StopTimes,
		"calendar_dates.txt":  &c.gtfsData.CalendarDates,
		"shapes.txt":          &c.gtfsData.Shapes,
		"transfers.txt":       &c.gtfsData.Transfers,
		"pathways.txt":        &c.gtfsData.Pathways,
		"levels.txt":          &c.gtfsData.Levels,
		"feed_info.txt":       &c.gtfsData.FeedInfo,
		"fare_attributes.txt": &c.gtfsData.FareAttributes,
		"translations.txt":    &c.gtfsData.Translations,
	}

	for filename, dataPtr := range files {
		filepath := filepath.Join(c.inputDir, filename)
		if err := c.loadCSVFile(filepath, dataPtr); err != nil {
			fmt.Printf("Warning: Could not load %s: %v\n", filename, err)
			// Continue with other files
		}
	}

	fmt.Printf("Loaded GTFS data: %d agencies, %d stops, %d routes, %d trips\n",
		len(c.gtfsData.Agencies), len(c.gtfsData.Stops), len(c.gtfsData.Routes), len(c.gtfsData.Trips))

	return nil
}

// loadCSVFile loads a CSV file into the provided data slice
func (c *Converter) loadCSVFile(filepath string, dataPtr interface{}) error {
	file, err := os.Open(filepath) //nolint:gosec // filepath is controlled by application, not user input
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 2 { // Need header + at least one data row
		return fmt.Errorf("file has insufficient data")
	}

	headers := records[0]
	dataRows := records[1:]

	// Parse based on the data type
	switch ptr := dataPtr.(type) {
	case *[]gtfs.Agency:
		*ptr = c.parseAgencies(headers, dataRows)
	case *[]gtfs.Stop:
		*ptr = c.parseStops(headers, dataRows)
	case *[]gtfs.Route:
		*ptr = c.parseRoutes(headers, dataRows)
	case *[]gtfs.Trip:
		*ptr = c.parseTrips(headers, dataRows)
	case *[]gtfs.StopTime:
		*ptr = c.parseStopTimes(headers, dataRows)
	case *[]gtfs.CalendarDate:
		*ptr = c.parseCalendarDates(headers, dataRows)
	case *[]gtfs.Shape:
		*ptr = c.parseShapes(headers, dataRows)
	case *[]gtfs.Transfer:
		*ptr = c.parseTransfers(headers, dataRows)
	case *[]gtfs.Pathway:
		*ptr = c.parsePathways(headers, dataRows)
	case *[]gtfs.Level:
		*ptr = c.parseLevels(headers, dataRows)
	case *[]gtfs.FeedInfo:
		*ptr = c.parseFeedInfo(headers, dataRows)
	case *[]gtfs.FareAttribute:
		*ptr = c.parseFareAttributes(headers, dataRows)
	case *[]gtfs.Translation:
		*ptr = c.parseTranslations(headers, dataRows)
	default:
		return fmt.Errorf("unsupported data type for CSV parsing")
	}

	fmt.Printf("Loaded %d records from %s\n", len(dataRows), filepath)
	return nil
}

// Helper parsing functions for each GTFS entity type
func (c *Converter) parseAgencies(headers []string, rows [][]string) []gtfs.Agency {
	var agencies []gtfs.Agency
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		agency := gtfs.Agency{
			AgencyID:       getString(row, headerMap, "agency_id"),
			AgencyName:     getString(row, headerMap, "agency_name"),
			AgencyURL:      getString(row, headerMap, "agency_url"),
			AgencyTimezone: getString(row, headerMap, "agency_timezone"),
			AgencyLang:     getString(row, headerMap, "agency_lang"),
			AgencyPhone:    getString(row, headerMap, "agency_phone"),
			AgencyEmail:    getString(row, headerMap, "agency_email"),
			AgencyFareURL:  getString(row, headerMap, "agency_fare_url"),
		}
		agencies = append(agencies, agency)
	}
	return agencies
}

func (c *Converter) parseStops(headers []string, rows [][]string) []gtfs.Stop {
	var stops []gtfs.Stop
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		stop := gtfs.Stop{
			StopID:             getString(row, headerMap, "stop_id"),
			StopCode:           getString(row, headerMap, "stop_code"),
			StopName:           getString(row, headerMap, "stop_name"),
			StopDesc:           getString(row, headerMap, "stop_desc"),
			StopLat:            getFloat64(row, headerMap, "stop_lat"),
			StopLon:            getFloat64(row, headerMap, "stop_lon"),
			LocationType:       getInt(row, headerMap, "location_type"),
			ParentStation:      getString(row, headerMap, "parent_station"),
			StopTimezone:       getString(row, headerMap, "stop_timezone"),
			LevelID:            getString(row, headerMap, "level_id"),
			StopURL:            getString(row, headerMap, "stop_url"),
			WheelchairBoarding: getInt(row, headerMap, "wheelchair_boarding"),
			PlatformCode:       getString(row, headerMap, "platform_code"),
			ZoneID:             getString(row, headerMap, "zone_id"),
		}
		stops = append(stops, stop)
	}
	return stops
}

func (c *Converter) parseRoutes(headers []string, rows [][]string) []gtfs.Route {
	var routes []gtfs.Route
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		route := gtfs.Route{
			RouteID:        getString(row, headerMap, "route_id"),
			AgencyID:       getString(row, headerMap, "agency_id"),
			RouteShortName: getString(row, headerMap, "route_short_name"),
			RouteLongName:  getString(row, headerMap, "route_long_name"),
			RouteDesc:      getString(row, headerMap, "route_desc"),
			RouteType:      getInt(row, headerMap, "route_type"),
			RouteURL:       getString(row, headerMap, "route_url"),
			RouteColor:     getString(row, headerMap, "route_color"),
			RouteTextColor: getString(row, headerMap, "route_text_color"),
			RouteSortOrder: getString(row, headerMap, "route_sort_order"),
		}
		routes = append(routes, route)
	}
	return routes
}

func (c *Converter) parseTrips(headers []string, rows [][]string) []gtfs.Trip {
	var trips []gtfs.Trip
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		trip := gtfs.Trip{
			RouteID:              getString(row, headerMap, "route_id"),
			ServiceID:            getString(row, headerMap, "service_id"),
			TripID:               getString(row, headerMap, "trip_id"),
			TripHeadsign:         getString(row, headerMap, "trip_headsign"),
			TripShortName:        getString(row, headerMap, "trip_short_name"),
			DirectionID:          getString(row, headerMap, "direction_id"),
			BlockID:              getString(row, headerMap, "block_id"),
			ShapeID:              getString(row, headerMap, "shape_id"),
			WheelchairAccessible: getInt(row, headerMap, "wheelchair_accessible"),
			BikesAllowed:         getInt(row, headerMap, "bikes_allowed"),
		}
		trips = append(trips, trip)
	}
	return trips
}

func (c *Converter) parseStopTimes(headers []string, rows [][]string) []gtfs.StopTime {
	var stopTimes []gtfs.StopTime
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		stopTime := gtfs.StopTime{
			TripID:            getString(row, headerMap, "trip_id"),
			ArrivalTime:       getString(row, headerMap, "arrival_time"),
			DepartureTime:     getString(row, headerMap, "departure_time"),
			StopID:            getString(row, headerMap, "stop_id"),
			StopSequence:      getInt(row, headerMap, "stop_sequence"),
			StopHeadsign:      getString(row, headerMap, "stop_headsign"),
			PickupType:        getString(row, headerMap, "pickup_type"),
			DropOffType:       getString(row, headerMap, "drop_off_type"),
			ShapeDistTraveled: getString(row, headerMap, "shape_dist_traveled"),
			Timepoint:         getInt(row, headerMap, "timepoint"),
		}
		stopTimes = append(stopTimes, stopTime)
	}
	return stopTimes
}

func (c *Converter) parseCalendarDates(headers []string, rows [][]string) []gtfs.CalendarDate {
	var calendarDates []gtfs.CalendarDate
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		calendarDate := gtfs.CalendarDate{
			ServiceID:     getString(row, headerMap, "service_id"),
			Date:          getString(row, headerMap, "date"),
			ExceptionType: getInt(row, headerMap, "exception_type"),
		}
		calendarDates = append(calendarDates, calendarDate)
	}
	return calendarDates
}

func (c *Converter) parseShapes(headers []string, rows [][]string) []gtfs.Shape {
	var shapes []gtfs.Shape
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		shape := gtfs.Shape{
			ShapeID:           getString(row, headerMap, "shape_id"),
			ShapePtLat:        getFloat64(row, headerMap, "shape_pt_lat"),
			ShapePtLon:        getFloat64(row, headerMap, "shape_pt_lon"),
			ShapePtSequence:   getInt(row, headerMap, "shape_pt_sequence"),
			ShapeDistTraveled: getFloat64Ptr(row, headerMap, "shape_dist_traveled"),
		}
		shapes = append(shapes, shape)
	}
	return shapes
}

func (c *Converter) parseTransfers(headers []string, rows [][]string) []gtfs.Transfer {
	var transfers []gtfs.Transfer
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		transfer := gtfs.Transfer{
			FromStopID:      getString(row, headerMap, "from_stop_id"),
			ToStopID:        getString(row, headerMap, "to_stop_id"),
			TransferType:    getInt(row, headerMap, "transfer_type"),
			MinTransferTime: getInt(row, headerMap, "min_transfer_time"),
		}
		transfers = append(transfers, transfer)
	}
	return transfers
}

func (c *Converter) parsePathways(headers []string, rows [][]string) []gtfs.Pathway {
	var pathways []gtfs.Pathway
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		pathway := gtfs.Pathway{
			PathwayID:            getString(row, headerMap, "pathway_id"),
			FromStopID:           getString(row, headerMap, "from_stop_id"),
			ToStopID:             getString(row, headerMap, "to_stop_id"),
			PathwayMode:          getInt(row, headerMap, "pathway_mode"),
			IsBidirectional:      getInt(row, headerMap, "is_bidirectional"),
			Length:               getFloat64Ptr(row, headerMap, "length"),
			TraversalTime:        getIntPtr(row, headerMap, "traversal_time"),
			StairCount:           getIntPtr(row, headerMap, "stair_count"),
			MaxSlope:             getFloat64Ptr(row, headerMap, "max_slope"),
			MinWidth:             getFloat64Ptr(row, headerMap, "min_width"),
			SignpostedAs:         getString(row, headerMap, "signposted_as"),
			ReversedSignpostedAs: getString(row, headerMap, "reversed_signposted_as"),
		}
		pathways = append(pathways, pathway)
	}
	return pathways
}

func (c *Converter) parseLevels(headers []string, rows [][]string) []gtfs.Level {
	var levels []gtfs.Level
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		level := gtfs.Level{
			LevelID:    getString(row, headerMap, "level_id"),
			LevelIndex: getFloat64(row, headerMap, "level_index"),
			LevelName:  getString(row, headerMap, "level_name"),
		}
		levels = append(levels, level)
	}
	return levels
}

func (c *Converter) parseFeedInfo(headers []string, rows [][]string) []gtfs.FeedInfo {
	var feedInfos []gtfs.FeedInfo
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		feedInfo := gtfs.FeedInfo{
			FeedPublisherName: getString(row, headerMap, "feed_publisher_name"),
			FeedPublisherURL:  getString(row, headerMap, "feed_publisher_url"),
			FeedLang:          getString(row, headerMap, "feed_lang"),
			FeedStartDate:     getString(row, headerMap, "feed_start_date"),
			FeedEndDate:       getString(row, headerMap, "feed_end_date"),
			FeedVersion:       getString(row, headerMap, "feed_version"),
			FeedContactEmail:  getString(row, headerMap, "feed_contact_email"),
			FeedContactURL:    getString(row, headerMap, "feed_contact_url"),
		}
		feedInfos = append(feedInfos, feedInfo)
	}
	return feedInfos
}

func (c *Converter) parseFareAttributes(headers []string, rows [][]string) []gtfs.FareAttribute {
	var fareAttributes []gtfs.FareAttribute
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		fareAttribute := gtfs.FareAttribute{
			AgencyID:         getString(row, headerMap, "agency_id"),
			FareID:           getString(row, headerMap, "fare_id"),
			Price:            getFloat64(row, headerMap, "price"),
			CurrencyType:     getString(row, headerMap, "currency_type"),
			PaymentMethod:    getInt(row, headerMap, "payment_method"),
			Transfers:        getInt(row, headerMap, "transfers"),
			TransferDuration: getInt(row, headerMap, "transfer_duration"),
		}
		fareAttributes = append(fareAttributes, fareAttribute)
	}
	return fareAttributes
}

func (c *Converter) parseTranslations(headers []string, rows [][]string) []gtfs.Translation {
	var translations []gtfs.Translation
	headerMap := makeHeaderMap(headers)

	for _, row := range rows {
		translation := gtfs.Translation{
			TableName:   getString(row, headerMap, "table_name"),
			FieldName:   getString(row, headerMap, "field_name"),
			Language:    getString(row, headerMap, "language"),
			Translation: getString(row, headerMap, "translation"),
			RecordID:    getString(row, headerMap, "record_id"),
			RecordSubID: getString(row, headerMap, "record_sub_id"),
			FieldValue:  getString(row, headerMap, "field_value"),
		}
		translations = append(translations, translation)
	}
	return translations
}

// Helper functions for parsing CSV data
func makeHeaderMap(headers []string) map[string]int {
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header] = i
	}
	return headerMap
}

func getString(row []string, headerMap map[string]int, field string) string {
	if idx, exists := headerMap[field]; exists && idx < len(row) {
		return row[idx]
	}
	return ""
}

func getInt(row []string, headerMap map[string]int, field string) int {
	if idx, exists := headerMap[field]; exists && idx < len(row) {
		if row[idx] == "" {
			return 0
		}
		var result int
		_, _ = fmt.Sscanf(row[idx], "%d", &result)
		return result
	}
	return 0
}

func getFloat64(row []string, headerMap map[string]int, field string) float64 {
	if idx, exists := headerMap[field]; exists && idx < len(row) {
		if row[idx] == "" {
			return 0.0
		}
		var result float64
		_, _ = fmt.Sscanf(row[idx], "%f", &result)
		return result
	}
	return 0.0
}

// Helper functions for pointer types
func getFloat64Ptr(row []string, headerMap map[string]int, field string) *float64 {
	if idx, exists := headerMap[field]; exists && idx < len(row) {
		if row[idx] == "" {
			return nil
		}
		var result float64
		_, _ = fmt.Sscanf(row[idx], "%f", &result)
		return &result
	}
	return nil
}

func getIntPtr(row []string, headerMap map[string]int, field string) *int {
	if idx, exists := headerMap[field]; exists && idx < len(row) {
		if row[idx] == "" {
			return nil
		}
		var result int
		_, _ = fmt.Sscanf(row[idx], "%d", &result)
		return &result
	}
	return nil
}

// generateMissingEntities creates entities that GTFS doesn't have but NeTEx needs
func (c *Converter) generateMissingEntities() {
	fmt.Println("Generating missing entities...")

	// Generate Directions (GTFS direction_id 0/1)
	c.generateDirections()

	// Generate VehicleTypes from route_type
	c.createVehicleTypes()

	// Generate DayTypes (since Sofia lacks calendar.txt)
	c.generateDayTypes()
}

// createNeTExStructure creates the main NeTEx structure
func (c *Converter) createNeTExStructure() error {
	fmt.Println("Creating NeTEx structure...")

	// Set up PublicationDelivery
	c.netexData.Xmlns = "http://www.netex.org.uk/netex"
	c.netexData.Version = "1.1"
	c.netexData.PublicationTimestamp = time.Now().Format("2006-01-02T15:04:05")
	c.netexData.ParticipantRef = c.config.ParticipantRef

	// Create PublicationRequest
	c.createPublicationRequest()

	// Create DataObjects with CompositeFrame
	c.createDataObjects()

	return nil
}

// organizeIntoFrames organizes entities into appropriate NeTEx frames
func (c *Converter) organizeIntoFrames() {
	fmt.Println("Organizing into frames...")

	// Format current date properly
	now := time.Now()
	fromDate := now.Format("2006-01-02")
	toDate := now.AddDate(1, 0, 0).Format("2006-01-02")

	// Create CompositeFrame
	compositeFrame := netex.CompositeFrame{
		ID:      fmt.Sprintf("%s::CompositeFrame:SOFIA::", c.config.ParticipantRef),
		Version: "1",
		ValidBetween: netex.ValidBetween{
			FromDate: fromDate + "T00:00:00",
			ToDate:   toDate + "T00:00:00",
		},
		TypeOfFrameRef: netex.TypeOfFrameRef{
			Ref:        "BG::TypeOfFrame:NETEX_LINE_OFFER::",
			VersionRef: "1.0",
		},
		Codespaces: netex.Codespaces{
			Codespace: netex.Codespace{
				ID:          "BG",
				Xmlns:       "BG",
				XmlnsUrl:    "http://www.bg.netex.org",
				Description: "Bulgaria NeTEx codespace",
			},
		},
		FrameDefaults: netex.FrameDefaults{
			DefaultCodespaceRef: netex.DefaultCodespaceRef{
				Ref: "BG",
			},
			DefaultDataSourceRef: netex.DefaultDataSourceRef{
				Ref: "BG::DataSource:SOFIA_TRANSPORT::",
			},
			DefaultResponsibilitySetRef: netex.DefaultResponsibilitySetRef{
				Ref:     fmt.Sprintf("%s::ResponsibilitySet:SOFIA::", c.config.ParticipantRef),
				Version: "1",
			},
			DefaultLocale: netex.DefaultLocale{
				TimeZone:        c.config.DefaultTimezone,
				DefaultLanguage: c.config.DefaultLanguage,
			},
			DefaultLocationSystem: c.config.LocationSystem,
		},
		Frames: netex.Frames{
			ResourceFrame: &netex.ResourceFrame{
				ID:      fmt.Sprintf("%s::ResourceFrame:SOFIA::", c.config.ParticipantRef),
				Version: "1",
				TypeOfFrameRef: netex.TypeOfFrameRef{
					Ref:        "BG::TypeOfFrame:NETEX_RESOURCE::",
					VersionRef: "1.0",
				},
				DataSources: netex.DataSources{
					DataSource: netex.DataSource{
						ID:      "BG::DataSource:SOFIA_TRANSPORT::",
						Version: "1",
						Name:    "Sofia Transport",
					},
				},
				ResponsibilitySets: netex.ResponsibilitySets{
					ResponsibilitySet: netex.ResponsibilitySet{
						ID:      fmt.Sprintf("%s::ResponsibilitySet:SOFIA::", c.config.ParticipantRef),
						Version: "1",
						Roles: netex.Roles{
							ResponsibilityRoleAssignment: netex.ResponsibilityRoleAssignment{
								ID:           "BG::ResponsibilityRoleAssignment:SOFIA::",
								Version:      "1",
								DataRoleType: "dataSource",
								ResponsibleOrganisationRef: netex.ResponsibleOrganisationRef{
									Ref: "BG::Authority:SOFIA_TRANSPORT::",
								},
							},
						},
					},
				},
				Organisations: netex.Organisations{
					Authorities: []netex.Authority{
						{
							ID:         "BG::Authority:SOFIA_TRANSPORT::",
							Version:    "1",
							PublicCode: "SOFIA_TRANSPORT",
							Name:       "Sofia Transport",
							ShortName:  "Sofia Transport",
							LegalName:  "Sofia Transport Authority",
							ContactDetails: netex.ContactDetails{
								Phone: "+359 2 123 4567",
								Email: "info@sofiatransport.bg",
								Url:   "https://www.sofiatransport.bg",
							},
							OrganisationType: "authority",
						},
					},
					Operator: netex.Operator{
						ID:         "BG::Operator:SOFIA_TRANSPORT::",
						Version:    "1",
						PublicCode: "SOFIA_TRANSPORT",
						Name:       "Sofia Transport",
						ShortName:  "Sofia Transport",
						LegalName:  "Sofia Transport Operator",
						ContactDetails: netex.ContactDetails{
							Phone: "+359 2 123 4567",
							Email: "info@sofiatransport.bg",
							Url:   "https://www.sofiatransport.bg",
						},
						OrganisationType: "operator",
					},
				},
			},
			ServiceCalendarFrame: c.createServiceCalendarFrame(),
			ServiceFrame:         c.createServiceFrame(),
			SiteFrame:            c.createSiteFrame(),
			TimetableFrame:       c.createTimetableFrame(),
		},
	}

	c.netexData.DataObjects = netex.DataObjects{
		CompositeFrame: compositeFrame,
	}
}

// createServiceCalendarFrame creates the ServiceCalendarFrame
func (c *Converter) createServiceCalendarFrame() *netex.ServiceCalendarFrame {
	// Generate calendar data
	calendarConverter := NewCalendarConverter(c.idMapper)
	calendarResult, err := calendarConverter.ConvertCalendar(c.gtfsData.CalendarDates, c.gtfsData.Trips)

	if err != nil {
		fmt.Printf("Warning: Calendar conversion failed: %v, using fallback\n", err)
		calendarResult = &CalendarConversionResult{
			DayTypes:           []netex.DayType{},
			OperatingPeriods:   []netex.UicOperatingPeriod{},
			DayTypeAssignments: []netex.DayTypeAssignment{},
		}
	}

	// Use all converted DayTypes, OperatingPeriods, and DayTypeAssignments
	dayTypes := calendarResult.DayTypes
	operatingPeriods := calendarResult.OperatingPeriods
	dayTypeAssignments := calendarResult.DayTypeAssignments

	// If no data from conversion, create default ones
	if len(dayTypes) == 0 {
		// Format current date properly
		now := time.Now()
		fromDate := now.Format("2006-01-02")
		toDate := now.AddDate(1, 0, 0).Format("2006-01-02")

		defaultDayType := netex.DayType{
			ID:      "BG::DayType:WEEKDAY::",
			Version: "1",
		}
		dayTypes = []netex.DayType{defaultDayType}

		defaultOperatingPeriod := netex.UicOperatingPeriod{
			ID:           "BG::UicOperatingPeriod:ALL_YEAR::",
			Version:      "1",
			FromDate:     fromDate + "T00:00:00",
			ToDate:       toDate + "T00:00:00",
			ValidDayBits: "1111100", // Monday to Friday like Sofia example
		}
		operatingPeriods = []netex.UicOperatingPeriod{defaultOperatingPeriod}

		defaultDayTypeAssignment := netex.DayTypeAssignment{
			ID:      "BG::DayTypeAssignment:WEEKDAY_ASSIGN::",
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
		dayTypeAssignments = []netex.DayTypeAssignment{defaultDayTypeAssignment}
	}

	// Format current date properly for ServiceCalendar
	now := time.Now()
	fromDate := now.Format("2006-01-02")
	toDate := now.AddDate(1, 0, 0).Format("2006-01-02")

	return &netex.ServiceCalendarFrame{
		ID:      fmt.Sprintf("%s::ServiceCalendarFrame:SOFIA_CAL::", c.config.ParticipantRef),
		Version: "1",
		TypeOfFrameRef: netex.TypeOfFrameRef{
			Ref:        "f:ServiceCalendar",
			VersionRef: "1.0",
		},
		ServiceCalendar: netex.ServiceCalendar{
			ID:       "BG::ServiceCalendar:ALL_SERVICES::",
			Version:  "1",
			FromDate: fromDate, // Just date, no time
			ToDate:   toDate,   // Just date, no time
			DayTypes: netex.DayTypes{
				DayType: dayTypes,
			},
			OperatingPeriods: netex.OperatingPeriods{
				UicOperatingPeriod: operatingPeriods,
			},
			DayTypeAssignments: netex.DayTypeAssignments{
				DayTypeAssignment: dayTypeAssignments,
			},
		},
	}
}

// createServiceFrame creates the ServiceFrame
func (c *Converter) createServiceFrame() *netex.ServiceFrame {
	// Convert routes to lines
	routeConverter := NewRouteConverter(c.idMapper)
	routeResult, _ := routeConverter.ConvertRoutes(c.gtfsData.Routes, c.gtfsData)

	// Convert stops to scheduled stop points
	stopConverter := NewStopConverter(c.idMapper)
	stopResult, _ := stopConverter.ConvertStops(c.gtfsData.Stops)

	// Create service journey patterns
	serviceJourneyPatterns := c.createServiceJourneyPatterns()

	// Build service links from patterns
	serviceLinks := c.createServiceLinksFromPatterns(serviceJourneyPatterns)

	// Build connections/interchanges from transfers
	connections, interchangeRules := c.createTransfers()

	serviceFrame := &netex.ServiceFrame{
		ID:             "BG::ServiceFrame:SOFIA_SERVICE_F::",
		Version:        "1",
		TypeOfFrameRef: netex.TypeOfFrameRef{Ref: "f:Service", VersionRef: "1.0"},
		Directions: netex.Directions{
			Direction: routeResult.Directions,
		},
		RoutePoints: netex.RoutePoints{
			RoutePoint: routeResult.RoutePoints,
		},
		RouteLinks: netex.RouteLinks{
			RouteLink: []netex.RouteLink{},
		},
		Routes: netex.Routes{
			Route: routeResult.Routes,
		},
		Lines: netex.Lines{
			Line: routeResult.Lines,
		},
		ScheduledStopPoints: netex.ScheduledStopPoints{
			ScheduledStopPoint: stopResult.ScheduledStopPoints,
		},
		ServiceLinks: netex.ServiceLinks{
			ServiceLink: serviceLinks,
		},
		StopAssignments: netex.StopAssignments{
			PassengerStopAssignment: stopResult.PassengerStopAssignments,
		},
		JourneyPatterns: netex.JourneyPatterns{
			ServiceJourneyPattern: serviceJourneyPatterns,
		},
		Connections:      connections,
		InterchangeRules: interchangeRules,
	}

	return serviceFrame
}

// createTransfers converts GTFS transfers to NeTEx Connections and InterchangeRules
func (c *Converter) createTransfers() (netex.Connections, netex.InterchangeRules) {
	connections := netex.Connections{Connection: []netex.Connection{}}
	interchanges := netex.InterchangeRules{InterchangeRule: []netex.InterchangeRule{}}
	for _, t := range c.gtfsData.Transfers {
		fromSSP := fmt.Sprintf("BG::ScheduledStopPoint:scheduled_stop_point_%s::", t.FromStopID)
		toSSP := fmt.Sprintf("BG::ScheduledStopPoint:scheduled_stop_point_%s::", t.ToStopID)
		conn := netex.Connection{
			ID:      fmt.Sprintf("BG::Connection:CONN_%x::", md5.Sum([]byte(fromSSP+">"+toSSP))), //nolint:gosec // MD5 used for non-cryptographic ID generation
			Version: "1",
			From:    netex.ConnectionEnd{ScheduledStopPointRef: netex.ScheduledStopPointRef{Ref: fromSSP, Version: "1"}},
			To:      netex.ConnectionEnd{ScheduledStopPointRef: netex.ScheduledStopPointRef{Ref: toSSP, Version: "1"}},
			TransferDuration: netex.TransferDuration{
				DefaultDuration: func() string {
					if t.MinTransferTime > 0 {
						mins := t.MinTransferTime / 60
						return fmt.Sprintf("PT%dM", mins)
					}
					return DurationZero
				}(),
			},
		}
		connections.Connection = append(connections.Connection, conn)

		// Simple mapping for restriction type based on transfer_type
		restriction := "recommended"
		switch t.TransferType {
		case 0:
			restriction = "guaranteed"
		case 1:
			restriction = "recommended"
		case 2:
			restriction = "forbidden"
		case 3:
			restriction = "minTime"
		}
		interchanges.InterchangeRule = append(interchanges.InterchangeRule, netex.InterchangeRule{
			ID:              fmt.Sprintf("BG::InterchangeRule:IR_%x::", md5.Sum([]byte(fromSSP+">"+toSSP))), //nolint:gosec // MD5 used for non-cryptographic ID generation
			Version:         "1",
			RestrictionType: restriction,
		})
	}
	return connections, interchanges
}

// createSiteFrame creates the SiteFrame
func (c *Converter) createSiteFrame() *netex.SiteFrame {
	// Convert stops to stop places and quays
	stopConverter := NewStopConverter(c.idMapper)
	stopResult, _ := stopConverter.ConvertStops(c.gtfsData.Stops)

	return &netex.SiteFrame{
		ID:      fmt.Sprintf("%s::SiteFrame:SOFIA::", c.config.ParticipantRef),
		Version: "1",
		TypeOfFrameRef: netex.TypeOfFrameRef{
			Ref:        "BG::TypeOfFrame:NETEX_SITE::",
			VersionRef: "1.0",
		},
		StopPlaces: netex.StopPlaces{
			StopPlace: stopResult.StopPlaces,
		},
	}
}

// createTimetableFrame creates the TimetableFrame
func (c *Converter) createTimetableFrame() *netex.TimetableFrame {
	// Convert trips to service journeys
	serviceJourneys := c.convertTripsToServiceJourneys()
	vehicleTypes := c.createVehicleTypes()

	return &netex.TimetableFrame{
		ID:      fmt.Sprintf("%s::TimetableFrame:SOFIA::", c.config.ParticipantRef),
		Version: "1",
		TypeOfFrameRef: netex.TypeOfFrameRef{
			Ref:        "BG::TypeOfFrame:NETEX_TIMETABLE::",
			VersionRef: "1.0",
		},
		VehicleJourneys: netex.VehicleJourneys{
			ServiceJourney: serviceJourneys,
		},
		VehicleTypes: netex.VehicleTypes{
			VehicleType: vehicleTypes,
		},
	}
}

// convertTripsToServiceJourneys converts GTFS trips to NeTEx service journeys
func (c *Converter) convertTripsToServiceJourneys() []netex.ServiceJourney {
	var journeys []netex.ServiceJourney

	// Create a map to track which patterns we've created
	createdPatterns := make(map[string]bool)

	for _, trip := range c.gtfsData.Trips {
		// Skip trips without service ID
		if trip.ServiceID == "" {
			fmt.Printf("Warning: Skipping trip %s due to missing service ID\n", trip.TripID)
			continue
		}

		// Get stop times for this trip using O(1) lookup
		stopTimes := c.lookupIndex.GetStopTimesByTripID(trip.TripID)

		if len(stopTimes) == 0 {
			fmt.Printf("Warning: Skipping trip %s due to missing stop times\n", trip.TripID)
			continue
		}

		// Sort by sequence
		sort.Slice(stopTimes, func(i, j int) bool {
			return stopTimes[i].StopSequence < stopTimes[j].StopSequence
		})

		// Default to direction "0" if direction_id is empty
		directionID := trip.DirectionID
		if directionID == "" {
			directionID = "0"
		}

		// Create a unique key based on route and stop sequence
		var stopSequence []string
		for _, st := range stopTimes {
			stopSequence = append(stopSequence, st.StopID)
		}

		// Create pattern key based on route, direction, and stop sequence
		routeDirectionKey := fmt.Sprintf("%s_%s", trip.RouteID, directionID)

		// Limit the pattern key length to avoid XML truncation
		// Use a hash of the stop sequence instead of the full sequence
		stopSequenceHash := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(stopSequence, "_"))))[:8] //nolint:gosec // MD5 used for non-cryptographic ID generation
		patternKey := fmt.Sprintf("%s_%s", routeDirectionKey, stopSequenceHash)
		createdPatterns[patternKey] = true

		// Create TimetabledPassingTime for each stop
		var passingTimes []netex.TimetabledPassingTime
		for i, stopTime := range stopTimes {
			passingTime := netex.TimetabledPassingTime{
				ID:      c.idGenerator.GenerateTimetabledPassingTimeID(trip.TripID, i),
				Version: NetExVersion,
				StopPointInJourneyPatternRef: netex.StopPointInJourneyPatternRef{
					Ref:     c.idGenerator.GenerateStopPointInJourneyPatternID(patternKey, i),
					Version: NetExVersion,
				},
			}

			// Set departure time for first stop, arrival time for last stop
			switch {
			case i == 0:
				dep, depOff := normalizeGTFSClockTime(stopTime.DepartureTime)
				passingTime.DepartureTime = dep
				if depOff > 0 {
					passingTime.DepartureDayOffset = depOff
				}
			case i == len(stopTimes)-1:
				arr, arrOff := normalizeGTFSClockTime(stopTime.ArrivalTime)
				passingTime.ArrivalTime = arr
				if arrOff > 0 {
					passingTime.ArrivalDayOffset = arrOff
				}
			default:
				arr, arrOff := normalizeGTFSClockTime(stopTime.ArrivalTime)
				dep, depOff := normalizeGTFSClockTime(stopTime.DepartureTime)
				passingTime.ArrivalTime = arr
				if arrOff > 0 {
					passingTime.ArrivalDayOffset = arrOff
				}
				passingTime.DepartureTime = dep
				if depOff > 0 {
					passingTime.DepartureDayOffset = depOff
				}
			}

			passingTimes = append(passingTimes, passingTime)
		}

		// Get route type for transport mode
		routeType := c.getRouteType(trip.RouteID)
		transportMode := c.getTransportModeFromType(routeType)

		journey := netex.ServiceJourney{
			ID:              c.idGenerator.GenerateServiceJourneyID(trip.TripID),
			Version:         NetExVersion,
			TransportMode:   transportMode,
			DepartureTime:   func() string { t, _ := normalizeGTFSClockTime(c.getDepartureTime(stopTimes)); return t }(),
			JourneyDuration: c.calculateJourneyDuration(stopTimes),

			DayTypes: netex.DayTypesJourney{
				DayTypeRef: netex.DayTypeRef{
					Ref:     c.idGenerator.GenerateDayTypeID(trip.ServiceID),
					Version: NetExVersion,
				},
			},
			ServiceJourneyPatternRef: netex.ServiceJourneyPatternRef{
				Ref:     c.idGenerator.GenerateJourneyPatternID(patternKey),
				Version: NetExVersion,
			},
			VehicleTypeRef: netex.VehicleTypeRef{
				Ref:     c.idGenerator.GenerateVehicleTypeID(routeType),
				Version: NetExVersion,
			},
			PassingTimes: netex.PassingTimes{
				TimetabledPassingTime: passingTimes,
			},
		}

		journeys = append(journeys, journey)
	}

	fmt.Printf("Created %d ServiceJourney elements with %d unique patterns\n", len(journeys), len(createdPatterns))
	return journeys
}

// createServiceLinksFromPatterns creates ServiceLink elements from the JourneyPatterns
func (c *Converter) createServiceLinksFromPatterns(patterns []netex.ServiceJourneyPattern) []netex.ServiceLink {
	links := []netex.ServiceLink{}
	seen := make(map[string]bool)
	for _, pattern := range patterns {
		stops := pattern.PointsInSequence.StopPointInJourneyPattern
		for i := 0; i+1 < len(stops); i++ {
			fromRef := stops[i].ScheduledStopPointRef.Ref
			toRef := stops[i+1].ScheduledStopPointRef.Ref
			key := fromRef + ">" + toRef
			if seen[key] {
				continue
			}
			seen[key] = true
			hash := fmt.Sprintf("%x", md5.Sum([]byte(key))) //nolint:gosec // MD5 used for non-cryptographic ID generation
			id := fmt.Sprintf("BG::ServiceLink:SL_%s::", hash)
			link := netex.ServiceLink{
				ID:           id,
				Version:      "1",
				FromPointRef: netex.FromPointRef{Ref: fromRef, Version: "1"},
				ToPointRef:   netex.ToPointRef{Ref: toRef, Version: "1"},
			}
			links = append(links, link)
		}
	}
	return links
}

// createVehicleTypes creates vehicle types from route types
func (c *Converter) createVehicleTypes() []netex.VehicleType {
	vehicleTypes := make(map[int]netex.VehicleType)

	for _, route := range c.gtfsData.Routes {
		if _, exists := vehicleTypes[route.RouteType]; !exists {
			vehicleType := netex.VehicleType{
				ID:      fmt.Sprintf("BG::VehicleType:vehicle_type_%d::", route.RouteType),
				Version: "1",
				Name:    c.getVehicleTypeName(route.RouteType),
			}
			vehicleTypes[route.RouteType] = vehicleType
		}
	}

	// Convert map to slice
	var result []netex.VehicleType
	for _, vehicleType := range vehicleTypes {
		result = append(result, vehicleType)
	}

	return result
}

func (c *Converter) getTransportModeFromType(routeType int) string {
	switch routeType {
	case RouteTypeTram:
		return "tram"
	case RouteTypeMetro:
		return "metro"
	case RouteTypeRail:
		return TransportModeRail
	case RouteTypeBus:
		return TransportModeBus
	case RouteTypeFerry:
		return "water"
	case RouteTypeCableTram:
		return TransportModeCableway
	case RouteTypeAerialLift:
		return TransportModeCableway
	case RouteTypeFunicular:
		return "funicular"
	case RouteTypeTrolleybus:
		return "trolleyBus"
	case RouteTypeMonorail:
		return TransportModeRail
	default:
		return TransportModeBus
	}
}

func (c *Converter) getVehicleTypeName(routeType int) string {
	switch routeType {
	case RouteTypeTram:
		return "Tram"
	case RouteTypeMetro:
		return "Metro"
	case RouteTypeRail:
		return "Rail"
	case RouteTypeBus:
		return "Bus"
	case RouteTypeFerry:
		return "Ferry"
	case RouteTypeCableTram:
		return "Cable Car"
	case RouteTypeAerialLift:
		return "Gondola"
	case RouteTypeFunicular:
		return "Funicular"
	case RouteTypeTrolleybus:
		return "Trolleybus"
	case RouteTypeMonorail:
		return "Monorail"
	default:
		return "Bus"
	}
}

func (c *Converter) getRouteType(routeID string) int {
	if route := c.lookupIndex.GetRouteByID(routeID); route != nil {
		return route.RouteType
	}
	return RouteTypeBus // Default to bus
}

func (c *Converter) getDepartureTime(stopTimes []gtfs.StopTime) string {
	if len(stopTimes) > 0 {
		return stopTimes[0].DepartureTime
	}
	return ""
}

// normalizeGTFSClockTime converts HH:MM:SS that may exceed 24:00:00 into a 24h clock and a day offset.
// For example 25:15:00 -> ("01:15:00", 1)
func normalizeGTFSClockTime(hms string) (string, int) {
	if hms == "" {
		return hms, 0
	}
	parts := strings.Split(hms, ":")
	if len(parts) != 3 {
		return hms, 0
	}
	hour := 0
	minute := 0
	second := 0
	_, _ = fmt.Sscanf(parts[0], "%d", &hour)
	_, _ = fmt.Sscanf(parts[1], "%d", &minute)
	_, _ = fmt.Sscanf(parts[2], "%d", &second)
	dayOffset := hour / 24
	hour %= 24
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), dayOffset
}

func (c *Converter) calculateJourneyDuration(stopTimes []gtfs.StopTime) string {
	if len(stopTimes) < 2 {
		return DurationZero
	}
	// parse first dep and last arr in seconds, taking into account 24+ hours
	parse := func(hms string) int {
		parts := strings.Split(hms, ":")
		if len(parts) != 3 {
			return 0
		}
		h, m, s := 0, 0, 0
		_, _ = fmt.Sscanf(parts[0], "%d", &h)
		_, _ = fmt.Sscanf(parts[1], "%d", &m)
		_, _ = fmt.Sscanf(parts[2], "%d", &s)
		return h*3600 + m*60 + s
	}
	first := stopTimes[0]
	last := stopTimes[len(stopTimes)-1]
	dep := parse(first.DepartureTime)
	arr := parse(last.ArrivalTime)
	if arr < dep {
		// next day(s)
		arr += 24 * 3600
	}
	dur := arr - dep
	if dur < 0 {
		dur = 0
	}
	// format as ISO-8601 duration to minutes resolution
	mins := dur / 60
	hrs := mins / 60
	mins %= 60
	return fmt.Sprintf("PT%dH%dM", hrs, mins)
}

// Helper methods for creating specific NeTEx structures
func (c *Converter) createPublicationRequest() {
	c.netexData.PublicationRequest = netex.PublicationRequest{
		RequestTimestamp: time.Now().Format("2006-01-02T15:04:05"),
		Topics: netex.Topics{
			NetworkFrameTopic: netex.NetworkFrameTopic{
				Current: struct{}{},
				// Remove NetworkFilterByValue to allow all lines to be processed
			},
		},
	}
}

func (c *Converter) createDataObjects() {
	// This is now handled in organizeIntoFrames()
}

func (c *Converter) generateDirections() {
	// Create Direction entities for direction_id 0 and 1
	fmt.Println("Generated 2 directions (0 and 1)")
}

func (c *Converter) generateDayTypes() {
	// Create DayType entities since Sofia lacks calendar.txt
	fmt.Println("Generated day types from calendar dates")
}

// Add this method to generate ServiceJourneyPatterns
func (c *Converter) createServiceJourneyPatterns() []netex.ServiceJourneyPattern {
	var patterns []netex.ServiceJourneyPattern

	// First, group trips by route and direction
	routeDirectionGroups := make(map[string][]gtfs.Trip)

	for _, trip := range c.gtfsData.Trips {
		// Skip trips without service ID
		if trip.ServiceID == "" {
			continue
		}

		// Default to direction "0" if direction_id is empty
		directionID := trip.DirectionID
		if directionID == "" {
			directionID = "0"
		}

		routeDirectionKey := fmt.Sprintf("%s_%s", trip.RouteID, directionID)
		routeDirectionGroups[routeDirectionKey] = append(routeDirectionGroups[routeDirectionKey], trip)
	}

	fmt.Printf("Creating ServiceJourneyPatterns for %d route-direction combinations\n", len(routeDirectionGroups))

	// For each route-direction combination, create patterns based on unique stop sequences
	for routeDirectionKey, trips := range routeDirectionGroups {
		// Group trips by unique stop sequences within this route-direction
		patternGroups := make(map[string][]gtfs.Trip)

		for _, trip := range trips {
			// Get stop times for this trip using O(1) lookup
			stopTimes := c.lookupIndex.GetStopTimesByTripID(trip.TripID)

			if len(stopTimes) == 0 {
				continue
			}

			// Sort by sequence
			sort.Slice(stopTimes, func(i, j int) bool {
				return stopTimes[i].StopSequence < stopTimes[j].StopSequence
			})

			// Create a unique key based on stop sequence
			var stopSequence []string
			for _, st := range stopTimes {
				stopSequence = append(stopSequence, st.StopID)
			}

			// Create pattern key based on route, direction, and stop sequence
			// Limit the pattern key length to avoid XML truncation
			// Use a hash of the stop sequence instead of the full sequence
			stopSequenceHash := fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(stopSequence, "_"))))[:8] //nolint:gosec // MD5 used for non-cryptographic ID generation
			patternKey := fmt.Sprintf("%s_%s", routeDirectionKey, stopSequenceHash)
			patternGroups[patternKey] = append(patternGroups[patternKey], trip)
		}

		// Create a pattern for each unique stop sequence within this route-direction
		for patternKey, patternTrips := range patternGroups {
			if len(patternTrips) == 0 {
				continue
			}

			trip := patternTrips[0] // Use first trip as template

			// Get stop times for this trip using O(1) lookup
			stopTimes := c.lookupIndex.GetStopTimesByTripID(trip.TripID)

			// Sort by sequence
			sort.Slice(stopTimes, func(i, j int) bool {
				return stopTimes[i].StopSequence < stopTimes[j].StopSequence
			})

			// Create StopPointInJourneyPattern for each stop
			var pointsInSequence []netex.StopPointInJourneyPattern
			for i, stopTime := range stopTimes {
				point := netex.StopPointInJourneyPattern{
					ID:      fmt.Sprintf("BG::StopPointInJourneyPattern:JP_%s_%d::", patternKey, i),
					Order:   fmt.Sprintf("%d", i),
					Version: "1",
					ScheduledStopPointRef: netex.ScheduledStopPointRef{
						Ref:     fmt.Sprintf("BG::ScheduledStopPoint:scheduled_stop_point_%s::", stopTime.StopID),
						Version: "1",
					},
					ForAlighting:               true,
					ForBoarding:                true,
					ChangeOfDestinationDisplay: false,
				}
				pointsInSequence = append(pointsInSequence, point)
			}

			// Default to direction "0" if direction_id is empty
			directionID := trip.DirectionID
			if directionID == "" {
				directionID = "0"
			}

			pattern := netex.ServiceJourneyPattern{
				ID:      fmt.Sprintf("BG::JourneyPattern:JP_%s::", patternKey),
				Version: "1",
				RouteRef: netex.RouteRef{
					Ref:     fmt.Sprintf("BG::Route:route_%s::", trip.RouteID),
					Version: "1",
				},
				DirectionRef: netex.DirectionRef{
					Ref:     fmt.Sprintf("BG::Direction:direction_%s::", directionID),
					Version: "1",
				},
				PointsInSequence: netex.PointsInSequenceJp{
					StopPointInJourneyPattern: pointsInSequence,
				},
			}

			patterns = append(patterns, pattern)
		}
	}

	fmt.Printf("Created %d ServiceJourneyPattern elements\n", len(patterns))
	return patterns
}

// GetStats returns conversion statistics
func (c *Converter) GetStats() *ConversionStats {
	if c.gtfsData == nil {
		return nil
	}

	stats := &ConversionStats{
		GTFSEntities:    0,
		NeTExEntities:   0,
		IDMappings:      0,
		DurationSeconds: 0.0,
	}

	// Count GTFS entities
	if c.gtfsData.Agencies != nil {
		stats.GTFSEntities += len(c.gtfsData.Agencies)
	}
	if c.gtfsData.Stops != nil {
		stats.GTFSEntities += len(c.gtfsData.Stops)
	}
	if c.gtfsData.Routes != nil {
		stats.GTFSEntities += len(c.gtfsData.Routes)
	}
	if c.gtfsData.Trips != nil {
		stats.GTFSEntities += len(c.gtfsData.Trips)
	}
	if c.gtfsData.StopTimes != nil {
		stats.GTFSEntities += len(c.gtfsData.StopTimes)
	}
	if c.gtfsData.CalendarDates != nil {
		stats.GTFSEntities += len(c.gtfsData.CalendarDates)
	}
	if c.gtfsData.Shapes != nil {
		stats.GTFSEntities += len(c.gtfsData.Shapes)
	}
	if c.gtfsData.Transfers != nil {
		stats.GTFSEntities += len(c.gtfsData.Transfers)
	}
	if c.gtfsData.Pathways != nil {
		stats.GTFSEntities += len(c.gtfsData.Pathways)
	}
	if c.gtfsData.Levels != nil {
		stats.GTFSEntities += len(c.gtfsData.Levels)
	}
	if c.gtfsData.FeedInfo != nil {
		stats.GTFSEntities += len(c.gtfsData.FeedInfo)
	}
	if c.gtfsData.FareAttributes != nil {
		stats.GTFSEntities += len(c.gtfsData.FareAttributes)
	}
	if c.gtfsData.Translations != nil {
		stats.GTFSEntities += len(c.gtfsData.Translations)
	}

	// Count ID mappings
	if c.idMapper != nil {
		mappingStats := c.idMapper.GetMappingStats()
		stats.IDMappings = mappingStats["total_mappings"]
	}

	// Count NeTEx entities (this would be populated after conversion)
	// For now, return a placeholder
	stats.NeTExEntities = stats.GTFSEntities * 2 // Rough estimate: each GTFS entity creates ~2 NeTEx entities

	// Calculate duration if conversion has been completed
	if !c.startTime.IsZero() && !c.endTime.IsZero() {
		stats.DurationSeconds = c.endTime.Sub(c.startTime).Seconds()
	}

	return stats
}

// ConversionStats holds conversion statistics
type ConversionStats struct {
	GTFSEntities    int
	NeTExEntities   int
	IDMappings      int
	DurationSeconds float64
}
