package converter

import (
	"fmt"
	"sort"
	"strconv"

	"gtfs-netex-converter/gtfs"
	"gtfs-netex-converter/netex"
)

// RouteConverter handles conversion of GTFS routes to NeTEx lines
type RouteConverter struct {
	idMapper *IDMapper
}

// NewRouteConverter creates a new route converter
func NewRouteConverter(idMapper *IDMapper) *RouteConverter {
	return &RouteConverter{
		idMapper: idMapper,
	}
}

// ConvertRoutes converts GTFS routes to NeTEx lines and directions
func (rc *RouteConverter) ConvertRoutes(gtfsRoutes []gtfs.Route, gtfsData *GTFSData) (*RouteConversionResult, error) {
	fmt.Println("Converting routes...")

	result := &RouteConversionResult{
		Lines:       []netex.Line{},
		Routes:      []netex.Route{},
		Directions:  []netex.Direction{},
		RoutePoints: []netex.RoutePoint{},
	}

	// Generate directions first (GTFS direction_id 0 and 1)
	result.Directions = rc.generateDirections()

	// Create a map to track which route points we've created
	createdRoutePoints := make(map[string]bool)

	// Convert each route to a line and route
	for _, route := range gtfsRoutes {
		line := rc.convertToLine(route)
		neTExRoute, routePoints := rc.convertToRoute(route, gtfsData)

		// Add route points that haven't been created yet
		for _, routePoint := range routePoints {
			if !createdRoutePoints[routePoint.ID] {
				result.RoutePoints = append(result.RoutePoints, routePoint)
				createdRoutePoints[routePoint.ID] = true
			}
		}

		result.Lines = append(result.Lines, line)
		result.Routes = append(result.Routes, neTExRoute)
	}

	fmt.Printf("Converted %d routes to %d lines, %d routes, %d directions, and %d route points\n",
		len(gtfsRoutes), len(result.Lines), len(result.Routes), len(result.Directions), len(result.RoutePoints))

	return result, nil
}

// RouteConversionResult holds the result of route conversion
type RouteConversionResult struct {
	Lines       []netex.Line
	Routes      []netex.Route
	Directions  []netex.Direction
	RoutePoints []netex.RoutePoint
}

// convertToLine converts a GTFS route to NeTEx Line
func (rc *RouteConverter) convertToLine(route gtfs.Route) netex.Line {
	lineID := fmt.Sprintf("BG::Line:line_%s::", route.RouteID)

	line := netex.Line{
		ID:            lineID,
		Version:       "1",
		Name:          route.RouteLongName,
		ShortName:     route.RouteShortName,
		TransportMode: rc.getTransportMode(route.RouteType),
		PublicCode:    route.RouteShortName,
		PrivateCode:   route.RouteID,
		AuthorityRef: netex.AuthorityRef{
			Ref:     "BG::Authority:SOFIA_TRANSPORT::",
			Version: "1",
		},
		KeyList: netex.KeyList{
			KeyValue: netex.KeyValue{
				Key:   "gtfs_route_id",
				Value: route.RouteID,
			},
		},
		AdditionalOperators: netex.AdditionalOperators{
			OperatorRef: netex.OperatorRef{
				Ref:     "BG::Operator:SOFIA_TRANSPORT::",
				Version: "1",
			},
		},
		Routes: netex.RouteRefs{
			RouteRef: []netex.RouteRef{
				{
					Ref:     fmt.Sprintf("BG::Route:route_%s::", route.RouteID),
					Version: "1",
				},
			},
		},
		AllowedDirections: netex.AllowedDirections{
			AllowedLineDirection: []netex.AllowedLineDirection{
				{
					ID:      fmt.Sprintf("allowed_direction_%s_0", route.RouteID),
					Version: "1",
					DirectionRef: netex.DirectionRef{
						Ref:     fmt.Sprintf("BG::Direction:direction_0::"),
						Version: "1",
					},
				},
				{
					ID:      fmt.Sprintf("allowed_direction_%s_1", route.RouteID),
					Version: "1",
					DirectionRef: netex.DirectionRef{
						Ref:     fmt.Sprintf("BG::Direction:direction_1::"),
						Version: "1",
					},
				},
			},
		},
	}

	// Add optional fields if present
	if route.RouteDesc != "" {
		// Note: NeTEx doesn't have a direct Description field for Line
		// Could be added to KeyList
	}

	if route.RouteURL != "" {
		// Note: NeTEx doesn't have a direct URL field for Line
		// Could be added to KeyList
	}

	if route.RouteColor != "" {
		// Note: NeTEx has Colour field but it's not in our Line struct
		// Could be added to KeyList
	}

	if route.RouteTextColor != "" {
		// Note: NeTEx has TextColour field but it's not in our Line struct
		// Could be added to KeyList
	}

	if route.RouteSortOrder != "" {
		// Note: NeTEx has Order field but it's not in our Line struct
		// Could be added to KeyList
	}

	return line
}

// generateDirections creates Direction entities for GTFS direction_id 0 and 1
func (rc *RouteConverter) generateDirections() []netex.Direction {
	directions := []netex.Direction{
		{
			ID:      fmt.Sprintf("BG::Direction:direction_0::"),
			Version: "1",
			// Note: NeTEx Direction doesn't have a Name field in our struct
			// Could be added to KeyList or extended struct
		},
		{
			ID:      fmt.Sprintf("BG::Direction:direction_1::"),
			Version: "1",
			// Note: NeTEx Direction doesn't have a Name field in our struct
			// Could be added to KeyList or extended struct
		},
	}

	return directions
}

// convertToRoute converts a GTFS route to NeTEx Route
func (rc *RouteConverter) convertToRoute(route gtfs.Route, gtfsData *GTFSData) (netex.Route, []netex.RoutePoint) {
	routeID := fmt.Sprintf("BG::Route:route_%s::", route.RouteID)

	// Find trips for this route
	var routeTrips []gtfs.Trip
	for _, trip := range gtfsData.Trips {
		if trip.RouteID == route.RouteID {
			routeTrips = append(routeTrips, trip)
		}
	}

	// Get unique stops for this route by analyzing stop times
	stopSequence := make(map[string]int)
	stopOrder := 0

	// Use the first trip to determine the route sequence
	if len(routeTrips) > 0 {
		firstTrip := routeTrips[0]
		for _, stopTime := range gtfsData.StopTimes {
			if stopTime.TripID == firstTrip.TripID {
				if _, exists := stopSequence[stopTime.StopID]; !exists {
					stopSequence[stopTime.StopID] = stopOrder
					stopOrder++
				}
			}
		}
	}

	// Create route points (RoutePoint elements)
	var routePoints []netex.RoutePoint
	for stopID, order := range stopSequence {
		// Find the stop to get location data
		var stopLat, stopLon float64
		for _, stop := range gtfsData.Stops {
			if stop.StopID == stopID {
				stopLat = stop.StopLat
				stopLon = stop.StopLon
				break
			}
		}

		routePoint := netex.RoutePoint{
			ID:      fmt.Sprintf("BG::RoutePoint:RPOINT_%s_%d::", route.RouteID, order),
			Version: "1",
			Location: netex.Location{
				Longitude: stopLon,
				Latitude:  stopLat,
			},
		}
		routePoints = append(routePoints, routePoint)
	}

	// Create PointOnRoute elements for the route's pointsInSequence
	var pointOnRoutes []netex.PointOnRoute
	for order := 0; order < len(stopSequence); order++ {
		pointOnRoute := netex.PointOnRoute{
			ID:      fmt.Sprintf("BG::PointOnRoute:POR_%s_%d::", route.RouteID, order),
			Order:   fmt.Sprintf("%d", order),
			Version: "1",
			RoutePointRef: netex.RoutePointRef{
				Ref:     fmt.Sprintf("BG::RoutePoint:RPOINT_%s_%d::", route.RouteID, order),
				Version: "1",
			},
		}
		pointOnRoutes = append(pointOnRoutes, pointOnRoute)
	}

	// Sort point on routes by order
	sort.Slice(pointOnRoutes, func(i, j int) bool {
		orderI, _ := strconv.Atoi(pointOnRoutes[i].Order)
		orderJ, _ := strconv.Atoi(pointOnRoutes[j].Order)
		return orderI < orderJ
	})

	neTExRoute := netex.Route{
		ID:      routeID,
		Version: "1",
		LineRef: netex.LineRef{
			Ref:     fmt.Sprintf("BG::Line:line_%s::", route.RouteID),
			Version: "1",
		},
		DirectionRef: netex.DirectionRef{
			Ref:     fmt.Sprintf("BG::Direction:direction_0::"), // Default direction
			Version: "1",
		},
		PointsInSequence: netex.PointsInSequence{
			PointOnRoute: pointOnRoutes,
		},
	}

	return neTExRoute, routePoints
}

// getTransportMode converts GTFS route_type to NeTEx TransportMode
func (rc *RouteConverter) getTransportMode(routeType int) string {
	switch routeType {
	case 0: // Tram
		return "tram"
	case 1: // Metro
		return "metro"
	case 2: // Rail
		return "rail"
	case 3: // Bus
		return "bus"
	case 4: // Ferry
		return "water"
	case 5: // Cable car
		return "cableway"
	case 6: // Gondola
		return "cableway"
	case 7: // Funicular
		return "funicular"
	case 11: // Trolleybus
		return "trolleyBus"
	case 12: // Monorail
		return "rail"
	default:
		return "bus"
	}
}
