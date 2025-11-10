package converter

import (
	"fmt"

	"gtfs-netex-converter/gtfs"
	"gtfs-netex-converter/netex"
)

// StopConverter handles conversion of GTFS stops to NeTEx stop entities
type StopConverter struct {
	idMapper *IDMapper
}

// NewStopConverter creates a new stop converter
func NewStopConverter(idMapper *IDMapper) *StopConverter {
	return &StopConverter{
		idMapper: idMapper,
	}
}

// ConvertStops converts GTFS stops to NeTEx stop entities
func (sc *StopConverter) ConvertStops(gtfsStops []gtfs.Stop) (*StopConversionResult, error) {
	fmt.Println("Converting stops...")

	result := &StopConversionResult{
		StopPlaces:               []netex.StopPlace{},
		ScheduledStopPoints:      []netex.ScheduledStopPoint{},
		Quays:                    []netex.Quay{},
		PassengerStopAssignments: []netex.PassengerStopAssignment{},
		Levels:                   []netex.Level{},
	}

	// Group stops by location type
	stopsByType := sc.groupStopsByType(gtfsStops)

	// Convert stations (location_type = 1) to StopPlaces
	for _, stop := range stopsByType[1] {
		stopPlace := sc.convertToStopPlace(stop)
		result.StopPlaces = append(result.StopPlaces, stopPlace)
	}

	// Convert regular stops (location_type = 0) to both StopPlace and ScheduledStopPoint
	for _, stop := range stopsByType[0] {
		// Create StopPlace for physical infrastructure
		stopPlace := sc.convertToStopPlace(stop)
		result.StopPlaces = append(result.StopPlaces, stopPlace)

		// Create ScheduledStopPoint for timetable
		scheduledStopPoint := sc.convertToScheduledStopPoint(stop)
		result.ScheduledStopPoints = append(result.ScheduledStopPoints, scheduledStopPoint)

		// Create Quay for platform
		quay := sc.convertToQuay(stop)
		result.Quays = append(result.Quays, quay)

		// Create PassengerStopAssignment to link them
		assignment := sc.createStopAssignment(stop, stopPlace, scheduledStopPoint, quay)
		result.PassengerStopAssignments = append(result.PassengerStopAssignments, assignment)

	}

	// Convert entrances (location_type = 2) - handled in infrastructure.go
	// Convert generic nodes (location_type = 3) - handled in infrastructure.go
	// Convert boarding areas (location_type = 4) - handled in infrastructure.go

	fmt.Printf("Converted %d stops to %d StopPlaces, %d ScheduledStopPoints, %d Quays, %d Assignments\n",
		len(gtfsStops), len(result.StopPlaces), len(result.ScheduledStopPoints), len(result.Quays), len(result.PassengerStopAssignments))

	return result, nil
}

// StopConversionResult holds the result of stop conversion
type StopConversionResult struct {
	StopPlaces               []netex.StopPlace
	ScheduledStopPoints      []netex.ScheduledStopPoint
	Quays                    []netex.Quay
	PassengerStopAssignments []netex.PassengerStopAssignment
	Levels                   []netex.Level
}

// groupStopsByType groups stops by their location_type
func (sc *StopConverter) groupStopsByType(stops []gtfs.Stop) map[int][]gtfs.Stop {
	grouped := make(map[int][]gtfs.Stop)

	for _, stop := range stops {
		grouped[stop.LocationType] = append(grouped[stop.LocationType], stop)
	}

	return grouped
}

// convertToStopPlace converts a GTFS stop to NeTEx StopPlace
func (sc *StopConverter) convertToStopPlace(stop gtfs.Stop) netex.StopPlace {
	stopPlaceID := fmt.Sprintf("BG::StopPlace:stop_place_%s::", stop.StopID)

	stopPlace := netex.StopPlace{
		ID:      stopPlaceID,
		Version: "1",
		Name:    stop.StopName,
		Centroid: netex.Centroid{
			Location: netex.Location{
				Longitude: stop.StopLon,
				Latitude:  stop.StopLat,
			},
		},
		StopPlaceType: sc.getStopPlaceType(stop.LocationType),
		// Note: Additional properties like wheelchair_boarding could be added to KeyList
	}

	// Note: Optional GTFS fields not mapped to NeTEx StopPlace:
	// - stop_desc: No direct Description field in our struct (could use KeyList)
	// - stop_url: No direct URL field in our struct (could use KeyList)
	// - zone_id: No direct ZoneID field in our struct (could use KeyList)

	return stopPlace
}

// convertToScheduledStopPoint converts a GTFS stop to NeTEx ScheduledStopPoint
func (sc *StopConverter) convertToScheduledStopPoint(stop gtfs.Stop) netex.ScheduledStopPoint {
	scheduledStopPointID := fmt.Sprintf("BG::ScheduledStopPoint:scheduled_stop_point_%s::", stop.StopID)

	scheduledStopPoint := netex.ScheduledStopPoint{
		ID:      scheduledStopPointID,
		Version: "1",
		Name:    stop.StopName,
		Location: netex.Location{
			Longitude: stop.StopLon,
			Latitude:  stop.StopLat,
		},
		PublicCode: stop.StopID,
		StopType:   sc.getStopType(stop.LocationType),
	}

	return scheduledStopPoint
}

// convertToQuay converts a GTFS stop to NeTEx Quay
func (sc *StopConverter) convertToQuay(stop gtfs.Stop) netex.Quay {
	quayID := fmt.Sprintf("BG::Quay:quay_%s::", stop.StopID)

	quay := netex.Quay{
		ID:      quayID,
		Version: "1",
		Name:    sc.generateQuayName(stop),
		Centroid: netex.Centroid{
			Location: netex.Location{
				Longitude: stop.StopLon,
				Latitude:  stop.StopLat,
			},
		},
		// Note: Additional properties could be added to KeyList
	}

	// Note: Optional GTFS fields not mapped to NeTEx Quay:
	// - platform_code: No direct PlatformCode field in our struct (could use KeyList)
	// - wheelchair_boarding: No direct WheelchairBoarding field in our struct (could use KeyList)

	return quay
}

// createStopAssignment creates a PassengerStopAssignment linking StopPlace, ScheduledStopPoint, and Quay
func (sc *StopConverter) createStopAssignment(stop gtfs.Stop, stopPlace netex.StopPlace, scheduledStopPoint netex.ScheduledStopPoint, quay netex.Quay) netex.PassengerStopAssignment {
	assignmentID := fmt.Sprintf("BG::PassengerStopAssignment:assignment_%s::", stop.StopID)

	assignment := netex.PassengerStopAssignment{
		ID:      assignmentID,
		Version: "1",
		Order:   "0",
		StopPlaceRef: netex.StopPlaceRef{
			Ref:     stopPlace.ID,
			Version: stopPlace.Version,
		},
		ScheduledStopPointRef: netex.ScheduledStopPointRef{
			Ref:     scheduledStopPoint.ID,
			Version: scheduledStopPoint.Version,
		},
		QuayRef: netex.QuayRef{
			Ref:     quay.ID,
			Version: quay.Version,
		},
	}

	return assignment
}

// getStopPlaceType converts GTFS location_type to NeTEx StopPlaceType
func (sc *StopConverter) getStopPlaceType(locationType int) string {
	switch locationType {
	case 0: // Stop/Platform
		return "busStop"
	case 1: // Station
		return "busStation"
	case 2: // Entrance/Exit
		return "entrance"
	case 3: // Generic Node
		return "accessSpace"
	case 4: // Boarding Area
		return "boardingPosition"
	default:
		return "busStop"
	}
}

// getStopType converts GTFS location_type to NeTEx StopType
func (sc *StopConverter) getStopType(locationType int) string {
	switch locationType {
	case 0: // Stop/Platform
		return "busStop"
	case 1: // Station
		return "busStation"
	case 2: // Entrance/Exit
		return "entrance"
	case 3: // Generic Node
		return "accessSpace"
	case 4: // Boarding Area
		return "boardingPosition"
	default:
		return "busStop"
	}
}

// generateQuayName generates a name for the quay
func (sc *StopConverter) generateQuayName(stop gtfs.Stop) string {
	if stop.PlatformCode != "" {
		return fmt.Sprintf("%s Platform %s", stop.StopName, stop.PlatformCode)
	}
	return fmt.Sprintf("%s Platform", stop.StopName)
}
