package netex

// ServiceCalendarFrame represents a ServiceCalendarFrame element.
type ServiceCalendarFrame struct {
	ID              string          `xml:"id,attr"`
	Version         string          `xml:"version,attr"`
	TypeOfFrameRef  TypeOfFrameRef  `xml:"TypeOfFrameRef"`
	ServiceCalendar ServiceCalendar `xml:"ServiceCalendar"`
}

// ServiceCalendar represents a ServiceCalendar element.
type ServiceCalendar struct {
	ID                 string             `xml:"id,attr"`
	Version            string             `xml:"version,attr"`
	FromDate           string             `xml:"FromDate"`
	ToDate             string             `xml:"ToDate"`
	DayTypes           DayTypes           `xml:"dayTypes"`
	OperatingPeriods   OperatingPeriods   `xml:"operatingPeriods"`
	DayTypeAssignments DayTypeAssignments `xml:"dayTypeAssignments"`
}

// DayTypes represents the dayTypes element.
type DayTypes struct {
	DayType []DayType `xml:"DayType"`
}

// DayType represents a DayType element.
type DayType struct {
	ID      string `xml:"id,attr"`
	Version string `xml:"version,attr"`
}

// OperatingPeriods represents the operatingPeriods element.
type OperatingPeriods struct {
	UicOperatingPeriod []UicOperatingPeriod `xml:"UicOperatingPeriod"`
}

// UicOperatingPeriod represents a UicOperatingPeriod element.
type UicOperatingPeriod struct {
	ID           string `xml:"id,attr"`
	Version      string `xml:"version,attr"`
	FromDate     string `xml:"FromDate"`
	ToDate       string `xml:"ToDate"`
	ValidDayBits string `xml:"ValidDayBits"`
}

// DayTypeAssignments represents the dayTypeAssignments element.
type DayTypeAssignments struct {
	DayTypeAssignment []DayTypeAssignment `xml:"DayTypeAssignment"`
}

// DayTypeAssignment represents a DayTypeAssignment element.
type DayTypeAssignment struct {
	ID                 string             `xml:"id,attr"`
	Order              string             `xml:"order,attr"`
	Version            string             `xml:"version,attr"`
	OperatingPeriodRef OperatingPeriodRef `xml:"OperatingPeriodRef"`
	DayTypeRef         DayTypeRef         `xml:"DayTypeRef"`
	IsAvailable        bool               `xml:"IsAvailable"`
}
