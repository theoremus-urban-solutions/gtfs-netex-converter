# GTFS Content the Converter Leaves Out

The `stream` converter writes only what EPIP allows (see [NeTEx Profile: EPIP](README.md#netex-profile-epip)). Everything it leaves out is counted in the conversion report's `warnings`, under the type in the last column.

## EPIP has no place for it

| GTFS | Why | Report warning |
|---|---|---|
| `trips.block_id` | EPIP has no `BlockRef` on `ServiceJourney` | `block_left_out` |
| `transfers` with `transfer_type` 3 | EPIP has no `TransferRestriction` | `transfer_not_possible_left_out` |
| `transfers` between routes (`from_route_id`, `to_route_id`) | EPIP has no `InterchangeRule` | `route_transfer_left_out` |
| the guarantee of a trip-to-trip `transfer_type` 1 | EPIP's `ServiceJourneyInterchange` has no `Guaranteed`; the interchange itself is kept | `guaranteed_left_out` |
| `trips.wheelchair_accessible` = 2 | EPIP's `MobilityFacilityList` has no "not suitable" value | `wheelchair_no_left_out` |
| `pathways.txt` | EPIP has no `SitePathLink` | `not_converted_pathways` |
| `fare_attributes.txt`, `fare_rules.txt` | fares are outside EPIP | `not_converted_fare_attributes`, `not_converted_fare_rules` |
| `stop_times.timepoint` | EPIP has `TimingPointStatus` only on the `ScheduledStopPoint`, not per call | none |

## Not converted yet

These fit EPIP and are planned for v1.2.0.

| GTFS | EPIP target | Report warning |
|---|---|---|
| `frequencies.txt` | `TemplateServiceJourney` with `HeadwayJourneyGroup`; until then such trips run only at their `stop_times` | `not_converted_frequencies` |
| `translations.txt` | `alternativeNames` on stops, `alternativeTexts` elsewhere | `not_converted_translations` |
| `levels.txt` | `Level` | `not_converted_levels` |
| entrances, generic nodes, boarding areas (`location_type` 2 to 4) | `StopPlaceEntrance`; nodes and boarding areas have no EPIP place | `stop_node_left_out` |
| `agency_timezone`, `agency_lang`, `feed_lang` | `FrameDefaults/DefaultLocale`; set through `Config` today | none |

## Other things the report counts

| Case | What happens | Report warning |
|---|---|---|
| routes with no trips | no `Line` is written, as nothing would reference it | `line_without_trips_left_out` |
| platforms no trip serves | `StopPlaceType` `other`, no `TransportMode` | `stop_without_trips` |
| shapes whose stops cannot be placed on them in order | their `ServiceLink`s carry no geometry | `shape_not_cut` |
| trips with no `stop_times` or no `service_id` | left out | `trip_without_stop_times` |
| trips whose `route_id` is not in `routes.txt` | left out | `trip_unknown_route_left_out` |
| trips that call at a `stop_id` that is not a platform in `stops.txt` | left out | `trip_unknown_stop_left_out` |
| trips whose first or last stop has no time | left out | `trip_without_end_times_left_out` |
| untimed stops between timed ones | times interpolated by distance between the timed stops around them | `times_interpolated` |
| platforms whose `parent_station` is missing or not a station | each becomes its own `StopPlace` | `parent_station_not_a_station` |
| transfers whose stop is not a platform or station | left out | `transfer_unknown_stop_left_out` |
| trip-to-trip transfers whose trip is not in the output, or whose stop the trip does not reach | left out | `interchange_unknown_trip_left_out`, `interchange_unknown_stop_left_out` |
| `route_type` with no EPIP mode (taxi, self-drive, miscellaneous, unknown codes) | written as `bus` | `route_type_unknown` |
| text that is not valid UTF-8, or holds control characters XML forbids | each bad character becomes U+FFFD | none |
