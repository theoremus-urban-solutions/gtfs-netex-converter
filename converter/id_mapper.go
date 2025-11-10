package converter

import (
	"fmt"
	"strings"
)

// IDMapper handles ID generation and reference tracking between GTFS and NeTEx
type IDMapper struct {
	gtfsToNetex map[string]string
	netexToGtfs map[string]string
	prefixes    map[string]string
}

// NewIDMapper creates a new ID mapper instance
func NewIDMapper() *IDMapper {
	return &IDMapper{
		gtfsToNetex: make(map[string]string),
		netexToGtfs: make(map[string]string),
		prefixes: map[string]string{
			"agency":             "operator_",
			"stop":               "stop_place_",
			"scheduled_stop":     "scheduled_stop_point_",
			"quay":               "quay_",
			"route":              "line_",
			"trip":               "service_journey_",
			"journey_pattern":    "journey_pattern_",
			"route_link":         "route_link_",
			"route_point":        "route_point_",
			"direction":          "direction_",
			"vehicle_type":       "vehicle_type_",
			"day_type":           "day_type_",
			"operating_period":   "operating_period_",
			"connection":         "connection_",
			"interchange_rule":   "interchange_rule_",
			"path_link":          "path_link_",
			"level":              "level_",
			"authority":          "authority_",
			"data_source":        "data_source_",
			"responsibility_set": "responsibility_set_",
			"tariff":             "tariff_",
		},
	}
}

// GenerateNeTExID generates a NeTEx ID from GTFS data
func (m *IDMapper) GenerateNeTExID(gtfsType, gtfsID string) string {
	key := fmt.Sprintf("%s:%s", gtfsType, gtfsID)

	if netexID, exists := m.gtfsToNetex[key]; exists {
		return netexID
	}

	prefix, ok := m.prefixes[gtfsType]
	if !ok {
		prefix = "entity_"
	}

	netexID := prefix + gtfsID
	m.gtfsToNetex[key] = netexID
	m.netexToGtfs[netexID] = key

	return netexID
}

// GetReference creates a reference structure for NeTEx
func (m *IDMapper) GetReference(gtfsType, gtfsID string) string {
	return m.GenerateNeTExID(gtfsType, gtfsID)
}

// GenerateStopPlaceID generates ID for StopPlace
func (m *IDMapper) GenerateStopPlaceID(stopID string) string {
	return m.GenerateNeTExID("stop", stopID)
}

// GenerateScheduledStopPointID generates ID for ScheduledStopPoint
func (m *IDMapper) GenerateScheduledStopPointID(stopID string) string {
	return m.GenerateNeTExID("scheduled_stop", stopID)
}

// GenerateQuayID generates ID for Quay
func (m *IDMapper) GenerateQuayID(stopID, platformCode string) string {
	if platformCode != "" {
		return m.GenerateNeTExID("quay", stopID+"_"+platformCode)
	}
	return m.GenerateNeTExID("quay", stopID+"_default")
}

// GenerateLineID generates ID for Line
func (m *IDMapper) GenerateLineID(routeID string) string {
	return m.GenerateNeTExID("route", routeID)
}

// GenerateServiceJourneyID generates ID for ServiceJourney
func (m *IDMapper) GenerateServiceJourneyID(tripID string) string {
	return m.GenerateNeTExID("trip", tripID)
}

// GenerateJourneyPatternID generates ID for ServiceJourneyPattern
func (m *IDMapper) GenerateJourneyPatternID(routeID, directionID string) string {
	patternKey := fmt.Sprintf("%s_%s", routeID, directionID)
	return m.GenerateNeTExID("journey_pattern", patternKey)
}

// GenerateRouteLinkID generates ID for RouteLink
func (m *IDMapper) GenerateRouteLinkID(shapeID string) string {
	return m.GenerateNeTExID("route_link", shapeID)
}

// GenerateRoutePointID generates ID for RoutePoint
func (m *IDMapper) GenerateRoutePointID(shapeID string, sequence int) string {
	pointKey := fmt.Sprintf("%s_%d", shapeID, sequence)
	return m.GenerateNeTExID("route_point", pointKey)
}

// GenerateDirectionID generates ID for Direction
func (m *IDMapper) GenerateDirectionID(directionID string) string {
	return m.GenerateNeTExID("direction", directionID)
}

// GenerateVehicleTypeID generates ID for VehicleType
func (m *IDMapper) GenerateVehicleTypeID(routeType int) string {
	vehicleTypeKey := fmt.Sprintf("type_%d", routeType)
	return m.GenerateNeTExID("vehicle_type", vehicleTypeKey)
}

// GenerateDayTypeID generates ID for DayType
func (m *IDMapper) GenerateDayTypeID(serviceID string) string {
	return m.GenerateNeTExID("day_type", serviceID)
}

// GenerateOperatingPeriodID generates ID for OperatingPeriod
func (m *IDMapper) GenerateOperatingPeriodID(serviceID string) string {
	return m.GenerateNeTExID("operating_period", serviceID)
}

// GenerateConnectionID generates ID for Connection
func (m *IDMapper) GenerateConnectionID(fromStopID, toStopID string) string {
	connectionKey := fmt.Sprintf("%s_to_%s", fromStopID, toStopID)
	return m.GenerateNeTExID("connection", connectionKey)
}

// GenerateInterchangeRuleID generates ID for InterchangeRule
func (m *IDMapper) GenerateInterchangeRuleID(fromStopID, toStopID string) string {
	ruleKey := fmt.Sprintf("%s_to_%s_rule", fromStopID, toStopID)
	return m.GenerateNeTExID("interchange_rule", ruleKey)
}

// GeneratePathLinkID generates ID for PathLink
func (m *IDMapper) GeneratePathLinkID(pathwayID string) string {
	return m.GenerateNeTExID("path_link", pathwayID)
}

// GenerateLevelID generates ID for Level
func (m *IDMapper) GenerateLevelID(levelID string) string {
	return m.GenerateNeTExID("level", levelID)
}

// GenerateAuthorityID generates ID for Authority
func (m *IDMapper) GenerateAuthorityID(authorityName string) string {
	// Clean authority name for ID
	cleanName := strings.ReplaceAll(authorityName, " ", "_")
	cleanName = strings.ToLower(cleanName)
	return m.GenerateNeTExID("authority", cleanName)
}

// GenerateDataSourceID generates ID for DataSource
func (m *IDMapper) GenerateDataSourceID(feedPublisherName string) string {
	cleanName := strings.ReplaceAll(feedPublisherName, " ", "_")
	cleanName = strings.ToLower(cleanName)
	return m.GenerateNeTExID("data_source", cleanName)
}

// GenerateResponsibilitySetID generates ID for ResponsibilitySet
func (m *IDMapper) GenerateResponsibilitySetID(participantRef string) string {
	return m.GenerateNeTExID("responsibility_set", participantRef)
}

// GenerateTariffID generates ID for Tariff
func (m *IDMapper) GenerateTariffID(fareID string) string {
	return m.GenerateNeTExID("tariff", fareID)
}

// GetMappingStats returns statistics about the ID mappings
func (m *IDMapper) GetMappingStats() map[string]int {
	return map[string]int{
		"total_mappings": len(m.gtfsToNetex),
		"gtfs_to_netex":  len(m.gtfsToNetex),
		"netex_to_gtfs":  len(m.netexToGtfs),
	}
}
