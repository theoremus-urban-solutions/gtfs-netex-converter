package netex

// Connection represents a Connection element for transfers
type Connection struct {
	ID                      string                `xml:"id,attr"`
	Version                 string                `xml:"version,attr"`
	From                    ConnectionEnd         `xml:"From"`
	To                      ConnectionEnd         `xml:"To"`
	TransferDuration        TransferDuration      `xml:"TransferDuration"`
}

// ConnectionEnd represents the From/To elements in a Connection
type ConnectionEnd struct {
	ScheduledStopPointRef ScheduledStopPointRef `xml:"ScheduledStopPointRef"`
}

// TransferDuration represents transfer duration
type TransferDuration struct {
	DefaultDuration string `xml:"DefaultDuration"`
}

// InterchangeRule represents an InterchangeRule element
type InterchangeRule struct {
	ID               string `xml:"id,attr"`
	Version          string `xml:"version,attr"`
	RestrictionType  string `xml:"RestrictionType"`
}

// Connections represents the connections element
type Connections struct {
	Connection []Connection `xml:"Connection"`
}

// InterchangeRules represents the interchangeRules element
type InterchangeRules struct {
	InterchangeRule []InterchangeRule `xml:"InterchangeRule"`
}