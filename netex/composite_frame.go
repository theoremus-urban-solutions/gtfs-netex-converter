package netex

// CompositeFrame represents a CompositeFrame element.
type CompositeFrame struct {
	ID             string         `xml:"id,attr"`
	Version        string         `xml:"version,attr"`
	ValidBetween   ValidBetween   `xml:"ValidBetween"`
	TypeOfFrameRef TypeOfFrameRef `xml:"TypeOfFrameRef"`
	Codespaces     Codespaces     `xml:"codespaces"`
	FrameDefaults  FrameDefaults  `xml:"FrameDefaults"`
	Frames         Frames         `xml:"frames"`
}

// ValidBetween represents the ValidBetween element.
type ValidBetween struct {
	FromDate string `xml:"FromDate"`
	ToDate   string `xml:"ToDate"`
}

// Codespaces represents the codespaces element.
type Codespaces struct {
	Codespace Codespace `xml:"Codespace"`
}

// Codespace represents a Codespace element.
type Codespace struct {
	ID          string `xml:"id,attr"`
	Xmlns       string `xml:"Xmlns"`
	XmlnsUrl    string `xml:"XmlnsUrl"`
	Description string `xml:"Description"`
}

// FrameDefaults represents the FrameDefaults element.
type FrameDefaults struct {
	DefaultCodespaceRef         DefaultCodespaceRef         `xml:"DefaultCodespaceRef"`
	DefaultDataSourceRef        DefaultDataSourceRef        `xml:"DefaultDataSourceRef"`
	DefaultResponsibilitySetRef DefaultResponsibilitySetRef `xml:"DefaultResponsibilitySetRef"`
	DefaultLocale               DefaultLocale               `xml:"DefaultLocale"`
	DefaultLocationSystem       string                      `xml:"DefaultLocationSystem"`
}

// DefaultLocale represents the DefaultLocale element.
type DefaultLocale struct {
	TimeZone        string `xml:"TimeZone"`
	DefaultLanguage string `xml:"DefaultLanguage"`
}

// Frames represents the frames element.
type Frames struct {
	GeneralFrame         *GeneralFrame         `xml:"GeneralFrame,omitempty"`
	ResourceFrame        *ResourceFrame        `xml:"ResourceFrame,omitempty"`
	ServiceCalendarFrame *ServiceCalendarFrame `xml:"ServiceCalendarFrame,omitempty"`
	ServiceFrame         *ServiceFrame         `xml:"ServiceFrame,omitempty"`
	SiteFrame            *SiteFrame            `xml:"SiteFrame,omitempty"`
	TimetableFrame       *TimetableFrame       `xml:"TimetableFrame,omitempty"`
	FareFrame            *FareFrame            `xml:"FareFrame,omitempty"`
}
