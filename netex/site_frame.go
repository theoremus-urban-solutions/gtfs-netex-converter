package netex

// SiteFrame represents a SiteFrame element.
type SiteFrame struct {
	ID                string            `xml:"id,attr"`
	Version           string            `xml:"version,attr"`
	TypeOfFrameRef    TypeOfFrameRef    `xml:"TypeOfFrameRef"`
	TopographicPlaces TopographicPlaces `xml:"topographicPlaces"`
	StopPlaces        StopPlaces        `xml:"stopPlaces"`
}

// TopographicPlaces represents the topographicPlaces element.
type TopographicPlaces struct {
	TopographicPlace []TopographicPlace `xml:"TopographicPlace"`
}

// TopographicPlace represents a TopographicPlace element.
type TopographicPlace struct {
	ID         string     `xml:"id,attr"`
	Version    string     `xml:"version,attr"`
	Name       string     `xml:"Name"`
	IsoCode    string     `xml:"IsoCode"`
	Descriptor Descriptor `xml:"Descriptor"`
}

// Descriptor represents a Descriptor element.
type Descriptor struct {
	Name string `xml:"Name"`
}

// StopPlaces represents the stopPlaces element.
type StopPlaces struct {
	StopPlace []StopPlace `xml:"StopPlace"`
}

// StopPlace represents a StopPlace element.
type StopPlace struct {
	ID                      string                  `xml:"id,attr"`
	Version                 string                  `xml:"version,attr"`
	KeyList                 KeyList                 `xml:"keyList"`
	Name                    string                  `xml:"Name"`
	Centroid                Centroid                `xml:"Centroid"`
	AlternativeNames        AlternativeNames        `xml:"alternativeNames"`
	TopographicPlaceRef     TopographicPlaceRef     `xml:"TopographicPlaceRef"`
	AuthorityRef            AuthorityRef            `xml:"AuthorityRef"`
	StopPlaceType           string                  `xml:"StopPlaceType"`
	AccessibilityAssessment AccessibilityAssessment `xml:"AccessibilityAssessment"`
	LevelRef                LevelRef                `xml:"LevelRef"`
	Url                     string                  `xml:"Url"`
	Quays                   Quays                   `xml:"quays"`
}

// Centroid represents a Centroid element.
type Centroid struct {
	Location Location `xml:"Location"`
}

// AlternativeNames represents the alternativeNames element.
type AlternativeNames struct {
	AlternativeName []AlternativeName `xml:"AlternativeName"`
}

// AlternativeName represents an AlternativeName element.
type AlternativeName struct {
	Name string `xml:"Name"`
}

// Quays represents the quays element.
type Quays struct {
	Quay []Quay `xml:"Quay"`
}

// Quay represents a Quay element.
type Quay struct {
	ID                      string                  `xml:"id,attr"`
	Version                 string                  `xml:"version,attr"`
	KeyList                 KeyList                 `xml:"keyList"`
	Name                    string                  `xml:"Name"`
	PublicCode              string                  `xml:"PublicCode"`
	Centroid                Centroid                `xml:"Centroid"`
	AccessibilityAssessment AccessibilityAssessment `xml:"AccessibilityAssessment"`
	LevelRef                LevelRef                `xml:"LevelRef"`
	Url                     string                  `xml:"Url"`
}
