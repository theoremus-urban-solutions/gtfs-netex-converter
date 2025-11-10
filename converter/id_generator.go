package converter

import "fmt"

// IDGenerator provides methods to generate NeTEx IDs with consistent formatting
type IDGenerator struct {
	participantRef string
}

// NewIDGenerator creates a new ID generator with the given participant reference
func NewIDGenerator(participantRef string) *IDGenerator {
	return &IDGenerator{
		participantRef: participantRef,
	}
}

// GenerateID generates a NeTEx ID with the format: ParticipantRef::Type:identifier::
func (g *IDGenerator) GenerateID(entityType, identifier string) string {
	return fmt.Sprintf("%s::%s:%s::", g.participantRef, entityType, identifier)
}

// GenerateStopPlaceID generates an ID for a StopPlace
func (g *IDGenerator) GenerateStopPlaceID(stopID string) string {
	return g.GenerateID("StopPlace", fmt.Sprintf("stop_place_%s", stopID))
}

// GenerateQuayID generates an ID for a Quay
func (g *IDGenerator) GenerateQuayID(stopID string) string {
	return g.GenerateID("Quay", fmt.Sprintf("quay_%s", stopID))
}

// GenerateScheduledStopPointID generates an ID for a ScheduledStopPoint
func (g *IDGenerator) GenerateScheduledStopPointID(stopID string) string {
	return g.GenerateID("ScheduledStopPoint", fmt.Sprintf("scheduled_stop_point_%s", stopID))
}

// GenerateRouteID generates an ID for a Route
func (g *IDGenerator) GenerateRouteID(routeID string) string {
	return g.GenerateID("Route", fmt.Sprintf("route_%s", routeID))
}

// GenerateLineID generates an ID for a Line
func (g *IDGenerator) GenerateLineID(routeID string) string {
	return g.GenerateID("Line", fmt.Sprintf("line_%s", routeID))
}

// GenerateServiceJourneyID generates an ID for a ServiceJourney
func (g *IDGenerator) GenerateServiceJourneyID(tripID string) string {
	return g.GenerateID("ServiceJourney", fmt.Sprintf("TRIP_%s", tripID))
}

// GenerateJourneyPatternID generates an ID for a JourneyPattern
func (g *IDGenerator) GenerateJourneyPatternID(patternKey string) string {
	return g.GenerateID("JourneyPattern", fmt.Sprintf("JP_%s", patternKey))
}

// GenerateStopPointInJourneyPatternID generates an ID for a StopPointInJourneyPattern
func (g *IDGenerator) GenerateStopPointInJourneyPatternID(patternKey string, order int) string {
	return g.GenerateID("StopPointInJourneyPattern", fmt.Sprintf("JP_%s_%d", patternKey, order))
}

// GenerateTimetabledPassingTimeID generates an ID for a TimetabledPassingTime
func (g *IDGenerator) GenerateTimetabledPassingTimeID(tripID string, order int) string {
	return g.GenerateID("TimetabledPassingTime", fmt.Sprintf("TPT_%s_%d", tripID, order))
}

// GenerateDayTypeID generates an ID for a DayType
func (g *IDGenerator) GenerateDayTypeID(serviceID string) string {
	return g.GenerateID("DayType", fmt.Sprintf("day_type_%s", serviceID))
}

// GenerateOperatingPeriodID generates an ID for an OperatingPeriod
func (g *IDGenerator) GenerateOperatingPeriodID(serviceID string) string {
	return g.GenerateID("OperatingPeriod", fmt.Sprintf("operating_period_%s", serviceID))
}

// GenerateDayTypeAssignmentID generates an ID for a DayTypeAssignment
func (g *IDGenerator) GenerateDayTypeAssignmentID(serviceID, date string) string {
	return g.GenerateID("DayTypeAssignment", fmt.Sprintf("day_type_assignment_%s_%s", serviceID, date))
}

// GenerateVehicleTypeID generates an ID for a VehicleType
func (g *IDGenerator) GenerateVehicleTypeID(routeType int) string {
	return g.GenerateID("VehicleType", fmt.Sprintf("vehicle_type_%d", routeType))
}

// GenerateServiceLinkID generates an ID for a ServiceLink
func (g *IDGenerator) GenerateServiceLinkID(hash string) string {
	return g.GenerateID("ServiceLink", fmt.Sprintf("SL_%s", hash))
}

// GenerateDirectionID generates an ID for a Direction
func (g *IDGenerator) GenerateDirectionID(directionID string) string {
	return g.GenerateID("Direction", fmt.Sprintf("direction_%s", directionID))
}

// GenerateInterchangeID generates an ID for an Interchange
func (g *IDGenerator) GenerateInterchangeID(fromStopID, toStopID string) string {
	return g.GenerateID("Interchange", fmt.Sprintf("interchange_%s_%s", fromStopID, toStopID))
}

// GenerateDataSourceID generates an ID for a DataSource
func (g *IDGenerator) GenerateDataSourceID(name string) string {
	return g.GenerateID("DataSource", name)
}

// GenerateResourceFrameID generates an ID for a ResourceFrame
func (g *IDGenerator) GenerateResourceFrameID(name string) string {
	return g.GenerateID("ResourceFrame", name)
}

// GenerateServiceFrameID generates an ID for a ServiceFrame
func (g *IDGenerator) GenerateServiceFrameID(name string) string {
	return g.GenerateID("ServiceFrame", name)
}

// GenerateServiceCalendarFrameID generates an ID for a ServiceCalendarFrame
func (g *IDGenerator) GenerateServiceCalendarFrameID(name string) string {
	return g.GenerateID("ServiceCalendarFrame", name)
}

// GenerateTimetableFrameID generates an ID for a TimetableFrame
func (g *IDGenerator) GenerateTimetableFrameID(name string) string {
	return g.GenerateID("TimetableFrame", name)
}

// GenerateAuthorityID generates an ID for an Authority
func (g *IDGenerator) GenerateAuthorityID(agencyID string) string {
	return g.GenerateID("Authority", fmt.Sprintf("authority_%s", agencyID))
}

// GenerateOperatorID generates an ID for an Operator
func (g *IDGenerator) GenerateOperatorID(agencyID string) string {
	return g.GenerateID("Operator", fmt.Sprintf("operator_%s", agencyID))
}
