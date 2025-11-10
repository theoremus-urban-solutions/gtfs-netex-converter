package netex

import "encoding/xml"

// ServiceFrame represents a ServiceFrame element.
type ServiceFrame struct {
	ID                  string              `xml:"id,attr"`
	Version             string              `xml:"version,attr"`
	TypeOfFrameRef      TypeOfFrameRef      `xml:"TypeOfFrameRef"`
	Directions          Directions          `xml:"directions"`
	RoutePoints         RoutePoints         `xml:"routePoints"`
	RouteLinks          RouteLinks          `xml:"routeLinks"`
	Routes              Routes              `xml:"routes"`
	Lines               Lines               `xml:"lines"`
	ScheduledStopPoints ScheduledStopPoints `xml:"scheduledStopPoints"`
	ServiceLinks        ServiceLinks        `xml:"serviceLinks"`
	StopAssignments     StopAssignments     `xml:"stopAssignments"`
	JourneyPatterns     JourneyPatterns     `xml:"journeyPatterns"`
	Connections         Connections         `xml:"connections"`
	InterchangeRules    InterchangeRules    `xml:"interchangeRules"`
}

// Directions represents the directions element.
type Directions struct {
	Direction []Direction `xml:"Direction"`
}

// Direction represents a Direction element.
type Direction struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
	Name    string `xml:"Name"`
}

// RoutePoints represents the routePoints element.
type RoutePoints struct {
	RoutePoint []RoutePoint `xml:"RoutePoint"`
}

// RoutePoint represents a RoutePoint element.
type RoutePoint struct {
	ID       string   `xml:"id,attr"`
	Version  string   `xml:"version,attr"`
	Location Location `xml:"Location"`
}

// Location represents a Location element.
type Location struct {
	Longitude float64 `xml:"Longitude"`
	Latitude  float64 `xml:"Latitude"`
}

// RouteLinks represents the routeLinks element.
type RouteLinks struct {
	RouteLink []RouteLink `xml:"RouteLink"`
}

// RouteLink represents a RouteLink element.
type RouteLink struct {
	ID           string       `xml:"id,attr"`
	Version      string       `xml:"version,attr"`
	Distance     int          `xml:"Distance"`
	LineString   LineString   `xml:"LineString"`
	FromPointRef FromPointRef `xml:"FromPointRef"`
	ToPointRef   ToPointRef   `xml:"ToPointRef"`
}

// LineString represents a LineString element.
type LineString struct {
	XMLName xml.Name `xml:"http://www.opengis.net/gml/3.2 LineString"`
	N0ID    string   `xml:"http://www.opengis.net/gml/3.2 id,attr"`
	Pos     []string `xml:"pos"`
}

// Routes represents the routes element.
type Routes struct {
	Route []Route `xml:"Route"`
}

// Route represents a Route element.
type Route struct {
	ID               string           `xml:"id,attr"`
	Version          string           `xml:"version,attr"`
	LineRef          LineRef          `xml:"LineRef"`
	DirectionRef     DirectionRef     `xml:"DirectionRef"`
	PointsInSequence PointsInSequence `xml:"pointsInSequence"`
}

// PointsInSequence represents the pointsInSequence element.
type PointsInSequence struct {
	PointOnRoute []PointOnRoute `xml:"PointOnRoute"`
}

// PointOnRoute represents a PointOnRoute element.
type PointOnRoute struct {
	ID            string        `xml:"id,attr"`
	Order         string        `xml:"order,attr"`
	Version       string        `xml:"version,attr"`
	RoutePointRef RoutePointRef `xml:"RoutePointRef"`
}

// ScheduledStopPoints represents the scheduledStopPoints element.
type ScheduledStopPoints struct {
	ScheduledStopPoint []ScheduledStopPoint `xml:"ScheduledStopPoint"`
}

// ScheduledStopPoint represents a ScheduledStopPoint element.
type ScheduledStopPoint struct {
	ID            string        `xml:"id,attr"`
	Version       string        `xml:"version,attr"`
	Name          string        `xml:"Name"`
	Location      Location      `xml:"Location"`
	PublicCode    string        `xml:"PublicCode"`
	StopType      string        `xml:"StopType"`
	TariffZoneRef TariffZoneRef `xml:"TariffZoneRef"`
}

// ServiceLinks represents the serviceLinks element.
type ServiceLinks struct {
	ServiceLink []ServiceLink `xml:"ServiceLink"`
}

// ServiceLink represents a ServiceLink element.
type ServiceLink struct {
	ID           string       `xml:"id,attr"`
	Version      string       `xml:"version,attr"`
	FromPointRef FromPointRef `xml:"FromPointRef"`
	ToPointRef   ToPointRef   `xml:"ToPointRef"`
}

// StopAssignments represents the stopAssignments element.
type StopAssignments struct {
	PassengerStopAssignment []PassengerStopAssignment `xml:"PassengerStopAssignment"`
}

// PassengerStopAssignment represents a PassengerStopAssignment element.
type PassengerStopAssignment struct {
	ID                    string                `xml:"id,attr"`
	Order                 string                `xml:"order,attr"`
	Version               string                `xml:"version,attr"`
	ScheduledStopPointRef ScheduledStopPointRef `xml:"ScheduledStopPointRef"`
	StopPlaceRef          StopPlaceRef          `xml:"StopPlaceRef"`
	QuayRef               QuayRef               `xml:"QuayRef"`
}

// JourneyPatterns represents the journeyPatterns element.
type JourneyPatterns struct {
	ServiceJourneyPattern []ServiceJourneyPattern `xml:"ServiceJourneyPattern"`
}

// ServiceJourneyPattern represents a ServiceJourneyPattern element.
type ServiceJourneyPattern struct {
	ID               string             `xml:"id,attr"`
	Version          string             `xml:"version,attr"`
	RouteRef         RouteRef           `xml:"RouteRef"`
	DirectionRef     DirectionRef       `xml:"DirectionRef"`
	PointsInSequence PointsInSequenceJp `xml:"pointsInSequence"`
}

// PointsInSequenceJp represents the pointsInSequence element within ServiceJourneyPattern.
type PointsInSequenceJp struct {
	StopPointInJourneyPattern []StopPointInJourneyPattern `xml:"StopPointInJourneyPattern"`
}

// StopPointInJourneyPattern represents a StopPointInJourneyPattern element.
type StopPointInJourneyPattern struct {
	ID                         string                `xml:"id,attr"`
	Order                      string                `xml:"order,attr"`
	Version                    string                `xml:"version,attr"`
	ScheduledStopPointRef      ScheduledStopPointRef `xml:"ScheduledStopPointRef"`
	ForAlighting               bool                  `xml:"ForAlighting"`
	ForBoarding                bool                  `xml:"ForBoarding"`
	ChangeOfDestinationDisplay bool                  `xml:"ChangeOfDestinationDisplay"`
	NoticeAssignments          NoticeAssignments     `xml:"noticeAssignments"`
}

// NoticeAssignments represents the noticeAssignments element.
type NoticeAssignments struct {
	NoticeAssignment []NoticeAssignment `xml:"NoticeAssignment"`
}

// NoticeAssignment represents a NoticeAssignment element.
type NoticeAssignment struct {
	ID        string    `xml:"id,attr"`
	Order     string    `xml:"order,attr"`
	Version   string    `xml:"version,attr"`
	Notice    Notice    `xml:"Notice"`
	NoticeRef NoticeRef `xml:"NoticeRef"`
}

// Notice represents a Notice element.
type Notice struct {
	ID         string `xml:"id,attr"`
	Version    string `xml:"version,attr"`
	Text       string `xml:"Text"`
	PublicCode string `xml:"PublicCode"`
	Lang       string `xml:"lang,attr"`
}
