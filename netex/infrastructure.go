package netex

// Level represents a Level element from GTFS levels
type Level struct {
	ID         string `xml:"id,attr"`
	Version    string `xml:"version,attr"`
	PublicCode string `xml:"PublicCode"`
	Name       string `xml:"Name"`
}

// Levels represents the levels element
type Levels struct {
	Level []Level `xml:"Level"`
}

// LevelRef represents a LevelRef element
type LevelRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// PathLink represents a PathLink element from GTFS pathways
type PathLink struct {
	ID                      string      `xml:"id,attr"`
	Version                 string      `xml:"version,attr"`
	From                    PlaceEnd    `xml:"From"`
	To                      PlaceEnd    `xml:"To"`
	TransitionType          string      `xml:"TransitionType"`
	BothWays                bool        `xml:"BothWays"`
	Distance                *float64    `xml:"Distance"`
	TransferDuration        TransferDuration `xml:"TransferDuration"`
	NumberOfSteps           *int        `xml:"NumberOfSteps"`
	MaximumGradient         *float64    `xml:"MaximumGradient"`
	MinimumWidth            *float64    `xml:"MinimumWidth"`
}

// PlaceEnd represents the From/To elements in a PathLink
type PlaceEnd struct {
	PlaceRef PlaceRef `xml:"PlaceRef"`
}

// PlaceRef represents a reference to a place (stop, entrance, etc.)
type PlaceRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}

// PathLinks represents the pathLinks element
type PathLinks struct {
	PathLink []PathLink `xml:"PathLink"`
}

// Entrance represents an Entrance element (GTFS location_type=2)
type Entrance struct {
	ID              string      `xml:"id,attr"`
	Version         string      `xml:"version,attr"`
	Name            string      `xml:"Name"`
	Centroid        Centroid    `xml:"Centroid"`
	ParentSiteRef   PlaceRef    `xml:"ParentSiteRef"`
}

// BoardingPosition represents a BoardingPosition element (GTFS location_type=4)
type BoardingPosition struct {
	ID              string      `xml:"id,attr"`
	Version         string      `xml:"version,attr"`
	Name            string      `xml:"Name"`
	Centroid        Centroid    `xml:"Centroid"`
	ParentQuayRef   QuayRef     `xml:"ParentQuayRef"`
}