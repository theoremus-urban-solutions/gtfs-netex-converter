package converter

// Constants for GTFS route types
const (
	RouteTypeTram       = 0
	RouteTypeMetro      = 1
	RouteTypeRail       = 2
	RouteTypeBus        = 3
	RouteTypeFerry      = 4
	RouteTypeCableTram  = 5
	RouteTypeAerialLift = 6
	RouteTypeFunicular  = 7
	RouteTypeTrolleybus = 11
	RouteTypeMonorail   = 12
)

// Constants for NeTEx versions and default values
const (
	NetExVersion      = "1"
	NetExNamespace    = "http://www.netex.org.uk/netex"
	DefaultDayPattern = "1111100" // Monday-Friday
	SecondsPerHour    = 3600
	SecondsPerMinute  = 60
	HoursPerDay       = 24
)

// Constants for time formats
const (
	ISO8601DateFormat = "2006-01-02"
	ISO8601TimeFormat = "15:04:05"
	NetExTimeFormat   = "15:04:05"
)

// Constants for ID hash lengths
const (
	IDHashLength = 8
)

// Constants for transport modes
const (
	TransportModeRail     = "rail"
	TransportModeBus      = "bus"
	TransportModeCableway = "cableway"
)

// Constants for stop place types
const (
	StopPlaceTypeBusStop = "busStop"
)

// Constants for durations
const (
	DurationZero = "PT0M"
)

// Constants for date formats
const (
	DateMin = "0000-01-01"
	DateMax = "9999-12-31"
)

// Constants for report formats
const (
	ReportFormatJSON = "json"
	ReportFormatText = "text"
	ReportFormatBoth = "both"
)
