package netex

// MultilingualString represents a string that can have multiple language versions
type MultilingualString struct {
	Value string `xml:",chardata"`
	Lang  string `xml:"lang,attr,omitempty"`
}

// AlternativeText represents alternative text in different languages
type AlternativeText struct {
	ID              string              `xml:"id,attr"`
	AttributeName   string              `xml:"AttributeName"`
	UseForLanguage  string              `xml:"UseForLanguage"`
	Text            MultilingualString  `xml:"Text"`
}

// AlternativeTexts represents a collection of alternative texts
type AlternativeTexts struct {
	AlternativeText []AlternativeText `xml:"AlternativeText"`
}