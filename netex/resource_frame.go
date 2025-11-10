package netex

// ResourceFrame represents a ResourceFrame element.
type ResourceFrame struct {
	ID                 string             `xml:"id,attr"`
	Version            string             `xml:"version,attr"`
	TypeOfFrameRef     TypeOfFrameRef     `xml:"TypeOfFrameRef"`
	DataSources        DataSources        `xml:"dataSources"`
	ResponsibilitySets ResponsibilitySets `xml:"responsibilitySets"`
	Organisations      Organisations      `xml:"organisations"`
}

// DataSources represents the dataSources element.
type DataSources struct {
	DataSource DataSource `xml:"DataSource"`
}

// DataSource represents a DataSource element.
type DataSource struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
	Name    string `xml:"Name"`
}

// ResponsibilitySets represents the responsibilitySets element.
type ResponsibilitySets struct {
	ResponsibilitySet ResponsibilitySet `xml:"ResponsibilitySet"`
}

// ResponsibilitySet represents a ResponsibilitySet element.
type ResponsibilitySet struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
	Roles   Roles  `xml:"roles"`
}

// Roles represents the roles element.
type Roles struct {
	ResponsibilityRoleAssignment ResponsibilityRoleAssignment `xml:"ResponsibilityRoleAssignment"`
}

// ResponsibilityRoleAssignment represents a ResponsibilityRoleAssignment element.
type ResponsibilityRoleAssignment struct {
	ID                         string                     `xml:"id,attr"`
	Version                    string                     `xml:"version,attr"`
	DataRoleType               string                     `xml:"DataRoleType"`
	ResponsibleOrganisationRef ResponsibleOrganisationRef `xml:"ResponsibleOrganisationRef"`
}

// Organisations represents the organisations element.
type Organisations struct {
	Authorities []Authority `xml:"Authority"`
	Operator    Operator    `xml:"Operator"`
}

// Authority represents an Authority element.
type Authority struct {
	ID               string         `xml:"id,attr"`
	Version          string         `xml:"version,attr"`
	PublicCode       string         `xml:"PublicCode"`
	Name             string         `xml:"Name"`
	ShortName        string         `xml:"ShortName"`
	LegalName        string         `xml:"LegalName"`
	ContactDetails   ContactDetails `xml:"ContactDetails"`
	OrganisationType string         `xml:"OrganisationType"`
}

// Operator represents an Operator element.
type Operator struct {
	ID               string         `xml:"id,attr"`
	Version          string         `xml:"version,attr"`
	PublicCode       string         `xml:"PublicCode"`
	Name             string         `xml:"Name"`
	ShortName        string         `xml:"ShortName"`
	LegalName        string         `xml:"LegalName"`
	ContactDetails   ContactDetails `xml:"ContactDetails"`
	OrganisationType string         `xml:"OrganisationType"`
}

// ContactDetails represents ContactDetails element.
type ContactDetails struct {
	Phone string `xml:"Phone"`
	Email string `xml:"Email"`
	Url   string `xml:"Url"`
}
