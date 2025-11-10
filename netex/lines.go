package netex

// Lines represents the lines element.
type Lines struct {
	Line []Line `xml:"Line"`
}

// Line represents a Line element.
type Line struct {
	ID                  string              `xml:"id,attr"`
	Version             string              `xml:"version,attr"`
	KeyList             KeyList             `xml:"keyList"`
	Name                string              `xml:"Name"`
	ShortName           string              `xml:"ShortName"`
	TransportMode       string              `xml:"TransportMode"`
	PublicCode          string              `xml:"PublicCode"`
	PrivateCode         string              `xml:"PrivateCode"`
	Colour              string              `xml:"Colour"`
	TextColour          string              `xml:"TextColour"`
	Order               string              `xml:"Order"`
	Url                 string              `xml:"Url"`
	Description         string              `xml:"Description"`
	AuthorityRef        AuthorityRef        `xml:"AuthorityRef"`
	AdditionalOperators AdditionalOperators `xml:"additionalOperators"`
	Routes              RouteRefs           `xml:"routes"`
	AllowedDirections   AllowedDirections   `xml:"allowedDirections"`
}

// KeyList represents a KeyList element.
type KeyList struct {
	KeyValue KeyValue `xml:"KeyValue"`
}

// KeyValue represents a KeyValue element.
type KeyValue struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

// AdditionalOperators represents the additionalOperators element.
type AdditionalOperators struct {
	OperatorRef OperatorRef `xml:"OperatorRef"`
}

// RouteRefs represents the routes element within Line.
type RouteRefs struct {
	RouteRef []RouteRef `xml:"RouteRef"`
}

// AllowedDirections represents the allowedDirections element.
type AllowedDirections struct {
	AllowedLineDirection []AllowedLineDirection `xml:"AllowedLineDirection"`
}

// AllowedLineDirection represents an AllowedLineDirection element.
type AllowedLineDirection struct {
	ID           string       `xml:"id,attr"`
	Version      string       `xml:"version,attr"`
	DirectionRef DirectionRef `xml:"DirectionRef"`
}
