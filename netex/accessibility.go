package netex

// AccessibilityAssessment represents accessibility information
type AccessibilityAssessment struct {
	ID                     string                   `xml:"id,attr"`
	Version                string                   `xml:"version,attr"`
	MobilityImpairedAccess string                   `xml:"MobilityImpairedAccess"`
	Limitations            AccessibilityLimitations `xml:"limitations"`
}

// AccessibilityLimitations represents accessibility limitations
type AccessibilityLimitations struct {
	AccessibilityLimitation AccessibilityLimitation `xml:"AccessibilityLimitation"`
}

// AccessibilityLimitation represents a specific accessibility limitation
type AccessibilityLimitation struct {
	WheelchairAccess string `xml:"WheelchairAccess"`
	StepFreeAccess   string `xml:"StepFreeAccess"`
}

// OnboardFacilities represents onboard facilities
type OnboardFacilities struct {
	BikeAllowed string `xml:"BikeAllowed"`
}
