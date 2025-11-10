package netex

// FareFrame represents a FareFrame element for fare structures
type FareFrame struct {
	ID             string         `xml:"id,attr"`
	Version        string         `xml:"version,attr"`
	TypeOfFrameRef TypeOfFrameRef `xml:"TypeOfFrameRef"`
	Tariffs        Tariffs        `xml:"tariffs"`
}

// Tariffs represents the tariffs element
type Tariffs struct {
	Tariff []Tariff `xml:"Tariff"`
}

// Tariff represents a Tariff element
type Tariff struct {
	ID                       string      `xml:"id,attr"`
	Version                  string      `xml:"version,attr"`
	Name                     string      `xml:"Name"`
	Amount                   float64     `xml:"Amount"`
	Currency                 string      `xml:"Currency"`
	PaymentMethod            string      `xml:"PaymentMethod"`
	MaximumNumberOfTransfers int         `xml:"MaximumNumberOfTransfers"`
	TransferDuration         string      `xml:"TransferDuration"`
	OperatorRef              OperatorRef `xml:"OperatorRef"`
}

// TariffZoneRef represents a TariffZoneRef element
type TariffZoneRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}
