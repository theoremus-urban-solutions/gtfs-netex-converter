package netex

// TimetableFrame represents a TimetableFrame element.
type TimetableFrame struct {
	ID              string          `xml:"id,attr"`
	Version         string          `xml:"version,attr"`
	TypeOfFrameRef  TypeOfFrameRef  `xml:"TypeOfFrameRef"`
	VehicleJourneys VehicleJourneys `xml:"vehicleJourneys"`
	VehicleTypes    VehicleTypes    `xml:"vehicleTypes"`
}

// VehicleJourneys represents the vehicleJourneys element.
type VehicleJourneys struct {
	ServiceJourney []ServiceJourney `xml:"ServiceJourney"`
}

// ServiceJourney represents a ServiceJourney element.
type ServiceJourney struct {
	ID              string `xml:"id,attr"`
	Version         string `xml:"version,attr"`
	TransportMode   string `xml:"TransportMode"`
	DepartureTime   string `xml:"DepartureTime"`
	JourneyDuration string `xml:"JourneyDuration"`

	DayTypes                 DayTypesJourney          `xml:"dayTypes"`
	ServiceJourneyPatternRef ServiceJourneyPatternRef `xml:"ServiceJourneyPatternRef"`
	VehicleTypeRef           VehicleTypeRef           `xml:"VehicleTypeRef"`
	AccessibilityAssessment  AccessibilityAssessment  `xml:"AccessibilityAssessment"`
	OnboardFacilities        OnboardFacilities        `xml:"OnboardFacilities"`
	PassingTimes             PassingTimes             `xml:"passingTimes"`
}

// DayTypesJourney represents the dayTypes element in ServiceJourney (contains DayTypeRef)
type DayTypesJourney struct {
	DayTypeRef DayTypeRef `xml:"DayTypeRef"`
}

// PassingTimes represents the passingTimes element.
type PassingTimes struct {
	TimetabledPassingTime []TimetabledPassingTime `xml:"TimetabledPassingTime"`
}

// TimetabledPassingTime represents a TimetabledPassingTime element.
type TimetabledPassingTime struct {
	ID                           string                       `xml:"id,attr"`
	Version                      string                       `xml:"version,attr"`
	StopPointInJourneyPatternRef StopPointInJourneyPatternRef `xml:"StopPointInJourneyPatternRef"`
	DepartureTime                string                       `xml:"DepartureTime"`
	ArrivalTime                  string                       `xml:"ArrivalTime"`
	DepartureDayOffset           int                          `xml:"DepartureDayOffset,omitempty"`
	ArrivalDayOffset             int                          `xml:"ArrivalDayOffset,omitempty"`
}

// VehicleTypes represents the vehicleTypes element.
type VehicleTypes struct {
	VehicleType []VehicleType `xml:"VehicleType"`
}

// VehicleType represents a VehicleType element.
type VehicleType struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
	Name    string `xml:"Name"`
}
