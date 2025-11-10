package converter

import "github.com/theoremus-urban-solutions/gtfs-netex-converter/gtfs"

// LookupIndices provides O(1) access to GTFS data
type LookupIndices struct {
	stopTimesByTripID  map[string][]gtfs.StopTime
	tripsByRouteID     map[string][]gtfs.Trip
	routesByID         map[string]*gtfs.Route
	stopsByID          map[string]*gtfs.Stop
	agenciesByID       map[string]*gtfs.Agency
	calendarDatesByID  map[string][]gtfs.CalendarDate
	transfersByStopID  map[string][]gtfs.Transfer
	shapesByID         map[string][]gtfs.Shape
	fareAttributesByID map[string]*gtfs.FareAttribute
}

// BuildLookupIndices builds all lookup indices from GTFS data
func BuildLookupIndices(data *GTFSData) *LookupIndices {
	indices := &LookupIndices{
		stopTimesByTripID:  make(map[string][]gtfs.StopTime),
		tripsByRouteID:     make(map[string][]gtfs.Trip),
		routesByID:         make(map[string]*gtfs.Route),
		stopsByID:          make(map[string]*gtfs.Stop),
		agenciesByID:       make(map[string]*gtfs.Agency),
		calendarDatesByID:  make(map[string][]gtfs.CalendarDate),
		transfersByStopID:  make(map[string][]gtfs.Transfer),
		shapesByID:         make(map[string][]gtfs.Shape),
		fareAttributesByID: make(map[string]*gtfs.FareAttribute),
	}

	// Index stop times by trip ID
	for i := range data.StopTimes {
		st := data.StopTimes[i]
		indices.stopTimesByTripID[st.TripID] = append(indices.stopTimesByTripID[st.TripID], st)
	}

	// Index trips by route ID
	for i := range data.Trips {
		trip := data.Trips[i]
		indices.tripsByRouteID[trip.RouteID] = append(indices.tripsByRouteID[trip.RouteID], trip)
	}

	// Index routes by ID
	for i := range data.Routes {
		route := &data.Routes[i]
		indices.routesByID[route.RouteID] = route
	}

	// Index stops by ID
	for i := range data.Stops {
		stop := &data.Stops[i]
		indices.stopsByID[stop.StopID] = stop
	}

	// Index agencies by ID
	for i := range data.Agencies {
		agency := &data.Agencies[i]
		indices.agenciesByID[agency.AgencyID] = agency
	}

	// Index calendar dates by service ID
	for i := range data.CalendarDates {
		cd := data.CalendarDates[i]
		indices.calendarDatesByID[cd.ServiceID] = append(indices.calendarDatesByID[cd.ServiceID], cd)
	}

	// Index transfers by stop ID
	for i := range data.Transfers {
		transfer := data.Transfers[i]
		indices.transfersByStopID[transfer.FromStopID] = append(indices.transfersByStopID[transfer.FromStopID], transfer)
	}

	// Index shapes by shape ID
	for i := range data.Shapes {
		shape := data.Shapes[i]
		indices.shapesByID[shape.ShapeID] = append(indices.shapesByID[shape.ShapeID], shape)
	}

	// Index fare attributes by ID
	for i := range data.FareAttributes {
		fare := &data.FareAttributes[i]
		indices.fareAttributesByID[fare.FareID] = fare
	}

	return indices
}

// GetStopTimesByTripID returns stop times for a given trip ID
func (idx *LookupIndices) GetStopTimesByTripID(tripID string) []gtfs.StopTime {
	return idx.stopTimesByTripID[tripID]
}

// GetTripsByRouteID returns trips for a given route ID
func (idx *LookupIndices) GetTripsByRouteID(routeID string) []gtfs.Trip {
	return idx.tripsByRouteID[routeID]
}

// GetRouteByID returns a route by its ID
func (idx *LookupIndices) GetRouteByID(routeID string) *gtfs.Route {
	return idx.routesByID[routeID]
}

// GetStopByID returns a stop by its ID
func (idx *LookupIndices) GetStopByID(stopID string) *gtfs.Stop {
	return idx.stopsByID[stopID]
}

// GetAgencyByID returns an agency by its ID
func (idx *LookupIndices) GetAgencyByID(agencyID string) *gtfs.Agency {
	return idx.agenciesByID[agencyID]
}

// GetCalendarDatesByServiceID returns calendar dates for a given service ID
func (idx *LookupIndices) GetCalendarDatesByServiceID(serviceID string) []gtfs.CalendarDate {
	return idx.calendarDatesByID[serviceID]
}

// GetTransfersByStopID returns transfers for a given stop ID
func (idx *LookupIndices) GetTransfersByStopID(stopID string) []gtfs.Transfer {
	return idx.transfersByStopID[stopID]
}

// GetShapesByShapeID returns shape points for a given shape ID
func (idx *LookupIndices) GetShapesByShapeID(shapeID string) []gtfs.Shape {
	return idx.shapesByID[shapeID]
}

// GetFareAttributeByID returns a fare attribute by its ID
func (idx *LookupIndices) GetFareAttributeByID(fareID string) *gtfs.FareAttribute {
	return idx.fareAttributesByID[fareID]
}
