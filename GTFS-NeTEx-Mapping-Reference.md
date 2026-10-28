# GTFS to NeTEx Mapping Reference

This is what the `stream` package (v1.1.0 and later) writes. The output follows EPIP and passes the strict EPIP schema and the NeTEx v2.0.0 schema; see [NeTEx Profile: EPIP](README.md#netex-profile-epip). Field choices follow the [DATA4PT GTFS-NeTEx mapping](./2024-June_DATA4PT_GTFS-NeTEx-Mapping_vf.pdf) wherever EPIP allows them.

What is left out, and why, is in [UNMAPPABLE_FEATURES.md](UNMAPPABLE_FEATURES.md).

## Frames

One `CompositeFrame` holds six frames. Each carries EPIP's frame type as `<TypeOfFrameRef ref="epip:…" versionRef="epip:1.0"/>`.

| Frame | EPIP type | Holds |
|---|---|---|
| CompositeFrame | `EU_PI_NETWORK_OFFER` | `ValidBetween` from `feed_info` dates (else the calendar's first and last day), codespace, `FrameDefaults` with time zone and language |
| ResourceFrame | `EU_PI_COMMON` | DataSource, ResponsibilitySet, Authority and Operator per agency |
| ServiceCalendarFrame | `EU_PI_CALENDAR` | DayType, UicOperatingPeriod, DayTypeAssignment |
| ServiceFrame | `EU_PI_NETWORK` | Direction, RoutePoint, Route, Line, DestinationDisplay, ScheduledStopPoint, ServiceLink, Connection, PassengerStopAssignment, ServiceJourneyPattern |
| SiteFrame | `EU_PI_STOP` | StopPlace with Quays |
| TimetableFrame | `EU_PI_TIMETABLE` | ServiceJourney, ServiceJourneyInterchange, VehicleType |

IDs have the form `<codespace>::<Type>:<local id>::`, for example `BG::StopPlace:stop_place_A0650::`. The codespace comes from `Config.Codespace` (CLI `-participant`).

## Files

### agency.txt

| GTFS | NeTEx |
|---|---|
| each row | one `Authority` and one `Operator` |
| `agency_id` (or `agency_name`) | `PublicCode`, and the key in their IDs |
| `agency_name` | `Name`, `ShortName`, `LegalName`; the first agency's name is also the `DataSource` name |
| `agency_email`, `agency_phone`, `agency_url` | `ContactDetails` `Email`, `Phone`, `Url` |

### routes.txt

A route becomes a `Line` only when at least one trip runs on it.

| GTFS | NeTEx `Line` |
|---|---|
| `route_id` | `keyList` key `gtfs_route_id`, and `PrivateCode` |
| `route_long_name` | `Name` |
| `route_short_name` | `ShortName` and `PublicCode` |
| `route_desc` | `Description` |
| `route_type` | `TransportMode` (0 tram, 1 metro, 2 rail, 3 bus, 4 water, 5 and 6 cableway, 7 funicular, 11 trolleyBus, 12 rail), and one `VehicleType` per type |
| `route_url` | `Url` |
| `route_color`, `route_text_color` | `Presentation` `Colour`, `TextColour` |
| `agency_id` | `AuthorityRef` and `additionalOperators/OperatorRef` |

Each route and `direction_id` pair becomes one `Route`, whose points follow its longest journey pattern.

### trips.txt and stop_times.txt

Trips that share route, direction, stops, shape, headsigns and pickup and drop-off types share one `ServiceJourneyPattern`.

| GTFS | NeTEx |
|---|---|
| trip | `ServiceJourney` |
| `trip_short_name` | `ServiceJourney/Name` (EPIP has no `PublicCode` here) |
| `service_id` | `dayTypes/DayTypeRef` |
| `trip_headsign` | `DestinationDisplay`, referenced from the journey pattern |
| `wheelchair_accessible` = 1 | `facilities/ServiceFacilitySet/MobilityFacilityList` `suitableForWheelchairs` |
| `bikes_allowed` = 1 or 2 | `facilities/ServiceFacilitySet/LuggageCarriageFacilityList` `cyclesAllowed` or `noCycles` |
| first `departure_time` | `DepartureTime` and `DepartureDayOffset` |
| last arrival minus first departure | `JourneyDuration` |
| `arrival_time`, `departure_time` | `TimetabledPassingTime` `ArrivalTime`, `DepartureTime`, with day offsets after midnight |
| `stop_id` in sequence | `StopPointInJourneyPattern/ScheduledStopPointRef` and `OnwardServiceLinkRef` |
| `pickup_type`, `drop_off_type` | `ForBoarding` false for 1, `ForAlighting` false for 1, `RequestStop` true for 2 or 3 |
| `stop_headsign` | `StopPointInJourneyPattern/DestinationDisplayRef` |

### stops.txt

| GTFS | NeTEx |
|---|---|
| platform (`location_type` 0) | `ScheduledStopPoint`, `Quay` and a `PassengerStopAssignment` linking them |
| station (`location_type` 1) | `StopPlace` holding its child platforms as `Quay`s |
| platform without `parent_station` | its own `StopPlace` with one `Quay` |
| `stop_name`, `stop_lat`, `stop_lon` | `Name` and `Location` / `Centroid` |
| `stop_code` | `ScheduledStopPoint/PublicCode` |
| `platform_code` | `Quay/PublicCode` |
| `stop_url` | `StopPlace/Url` |
| `wheelchair_boarding` 1 or 2 | `Quay/AccessibilityAssessment` with `WheelchairAccess` true or false |
| mode of the trips that serve it | `StopPlace/TransportMode` and `StopPlaceType` (`onstreetBus`, `onstreetTram`, `metroStation`, `railStation`, `busStation` for stations, …); `other` with no mode when no trip serves it |

### calendar.txt and calendar_dates.txt

| GTFS | NeTEx |
|---|---|
| `service_id` | `DayType` |
| weekday columns | `DayType/properties/PropertyOfDay/DaysOfWeek` |
| `start_date`..`end_date` | `UicOperatingPeriod` with `ValidDayBits`: one character per day, `1` on every day the service runs, holidays included |
| each `calendar_dates` row | a dated `DayTypeAssignment` with `isAvailable` true (type 1) or false (type 2) |

A service only in `calendar_dates.txt` gets the weekday pattern most of its days follow, plus dated exceptions, when that is shorter than listing every date. Otherwise every date is listed.

### shapes.txt

Each trip's shape is cut at its stops. Each stop is placed on the shape, never moving backwards, with the least total distance. Each stop-to-stop piece becomes the `gml:LineString` of a `ServiceLink`, shared by every pattern with the same stop pair and shape. When the stops cannot be placed in order, that pattern's links carry no geometry and the report counts it.

### transfers.txt

| GTFS | NeTEx |
|---|---|
| stop to stop, `transfer_type` 0, 1 or 2 | `Connection` with `WalkTransferDuration/DefaultDuration` = `min_transfer_time` seconds |
| trip to trip | `ServiceJourneyInterchange` |
| trip to trip, `transfer_type` 2 | `StandardTransferTime` = `min_transfer_time` |
| trip to trip, `transfer_type` 4 or 5 | `StaySeated` true or false |
| `transfer_type` 3, route to route, the guarantee of type 1 | left out, see [UNMAPPABLE_FEATURES.md](UNMAPPABLE_FEATURES.md) |
