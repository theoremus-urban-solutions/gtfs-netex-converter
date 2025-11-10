package netex

// Common reference types used throughout NeTEx

// TypeOfFrameRef represents a TypeOfFrameRef element.
type TypeOfFrameRef struct {
	Ref        string `xml:"ref,attr"`
	VersionRef string `xml:"versionRef,attr"`
}

// DefaultCodespaceRef represents a DefaultCodespaceRef element.
type DefaultCodespaceRef struct {
	Ref string `xml:"ref,attr"`
}

// DefaultDataSourceRef represents a DefaultDataSourceRef element.
type DefaultDataSourceRef struct {
	Ref string `xml:"ref,attr"`
}

// DefaultResponsibilitySetRef represents a DefaultResponsibilitySetRef element.
type DefaultResponsibilitySetRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// LineRef represents a LineRef element.
type LineRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// DirectionRef represents a DirectionRef element.
type DirectionRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// RouteRef represents a RouteRef element.
type RouteRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// AuthorityRef represents an AuthorityRef element.
type AuthorityRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// OperatorRef represents an OperatorRef element.
type OperatorRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// ResponsibleOrganisationRef represents a ResponsibleOrganisationRef element.
type ResponsibleOrganisationRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// ScheduledStopPointRef represents a ScheduledStopPointRef element.
type ScheduledStopPointRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// StopPlaceRef represents a StopPlaceRef element.
type StopPlaceRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// QuayRef represents a QuayRef element.
type QuayRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// RoutePointRef represents a RoutePointRef element.
type RoutePointRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// FromPointRef represents a FromPointRef element.
type FromPointRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// ToPointRef represents a ToPointRef element.
type ToPointRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// ServiceJourneyPatternRef represents a ServiceJourneyPatternRef element.
type ServiceJourneyPatternRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// VehicleTypeRef represents a VehicleTypeRef element.
type VehicleTypeRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// StopPointInJourneyPatternRef represents a StopPointInJourneyPatternRef element.
type StopPointInJourneyPatternRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// TopographicPlaceRef represents a TopographicPlaceRef element.
type TopographicPlaceRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// DayTypeRef represents a DayTypeRef element.
type DayTypeRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// OperatingPeriodRef represents an OperatingPeriodRef element.
type OperatingPeriodRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// NoticeRef represents a NoticeRef element.
type NoticeRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}
