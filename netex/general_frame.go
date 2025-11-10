package netex

// GeneralFrame represents a GeneralFrame element for infrastructure
type GeneralFrame struct {
	ID             string         `xml:"id,attr"`
	Version        string         `xml:"version,attr"`
	TypeOfFrameRef TypeOfFrameRef `xml:"TypeOfFrameRef"`
	Levels         Levels         `xml:"levels,omitempty"`
	PathLinks      PathLinks      `xml:"pathLinks,omitempty"`
}

// DestinationDisplay represents a DestinationDisplay element
type DestinationDisplay struct {
	ID              string `xml:"id,attr"`
	Version         string `xml:"version,attr"`
	Name            string `xml:"Name"`
	ShortName       string `xml:"ShortName"`
	SideText        string `xml:"SideText"`
	FrontText       string `xml:"FrontText"`
}

// DestinationDisplays represents the destinationDisplays element
type DestinationDisplays struct {
	DestinationDisplay []DestinationDisplay `xml:"DestinationDisplay"`
}

// DestinationDisplayRef represents a DestinationDisplayRef element
type DestinationDisplayRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// Block represents a Block element (vehicle block)
type Block struct {
	ID              string          `xml:"id,attr"`
	Version         string          `xml:"version,attr"`
	Name            string          `xml:"Name"`
	Description     string          `xml:"Description"`
	JourneyRefs     JourneyRefs     `xml:"journeys"`
}

// JourneyRefs represents journey references within a block
type JourneyRefs struct {
	ServiceJourneyRef []ServiceJourneyRef `xml:"ServiceJourneyRef"`
}

// ServiceJourneyRef represents a ServiceJourneyRef element
type ServiceJourneyRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// BlockRef represents a BlockRef element
type BlockRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}