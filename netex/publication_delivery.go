package netex

import "encoding/xml"

// PublicationDelivery represents the root element of the NeTEx XML.
type PublicationDelivery struct {
	XMLName              xml.Name           `xml:"PublicationDelivery"`
	Xmlns                string             `xml:"xmlns,attr"`
	XmlnsGml             string             `xml:"xmlns:gml,attr,omitempty"`
	Version              string             `xml:"version,attr"`
	PublicationTimestamp string             `xml:"PublicationTimestamp"`
	ParticipantRef       string             `xml:"ParticipantRef"`
	PublicationRequest   PublicationRequest `xml:"PublicationRequest"`
	DataObjects          DataObjects        `xml:"dataObjects"`
}

// PublicationRequest represents the PublicationRequest element.
type PublicationRequest struct {
	RequestTimestamp string `xml:"RequestTimestamp"`
	Topics           Topics `xml:"topics"`
}

// Topics represents the topics element within PublicationRequest.
type Topics struct {
	NetworkFrameTopic NetworkFrameTopic `xml:"NetworkFrameTopic"`
}

// NetworkFrameTopic represents the NetworkFrameTopic element.
type NetworkFrameTopic struct {
	Current              struct{}              `xml:"Current"`
	NetworkFilterByValue *NetworkFilterByValue `xml:"NetworkFilterByValue,omitempty"`
}

// NetworkFilterByValue represents the NetworkFilterByValue element.
type NetworkFilterByValue struct {
	ObjectReferences ObjectReferences `xml:"objectReferences"`
}

// ObjectReferences represents the objectReferences element.
type ObjectReferences struct {
	LineRef        LineRef        `xml:"LineRef"`
	TypeOfFrameRef TypeOfFrameRef `xml:"TypeOfFrameRef"`
}

// DataObjects represents the dataObjects element.
type DataObjects struct {
	CompositeFrame CompositeFrame `xml:"CompositeFrame"`
}
