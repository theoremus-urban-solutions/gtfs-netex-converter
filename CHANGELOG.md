# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - Unreleased

### Added
- `stream.ConvertZip`, `stream.ConvertZipFile` and `stream.ConvertFS`: a streaming converter whose output passes the strict EPIP schema and the NeTEx v2.0.0 schema. Full Sofia converts in about 5 s with under 200 MB of memory.
- Calendars: `UicOperatingPeriod` with `ValidDayBits` for every service, next to `DaysOfWeek` and dated `DayTypeAssignment`s. Every GTFS service day is kept, holidays included.
- Shapes: each trip's shape is cut at its stops into `ServiceLink` geometry, and each `StopPointInJourneyPattern` points to its onward link.
- Trip `wheelchair_accessible` and `bikes_allowed` as an inline `ServiceFacilitySet`; `trip_short_name` as the journey's `Name`.
- Report warnings that count everything EPIP cannot carry: blocks, pathways, fares, "transfer not possible", route-level transfers and timed-transfer guarantees.
- CLI: `-input` takes a GTFS zip; `-output` ending in `.zip` writes a zip; `-publication-time` fixes the timestamp.
- CI job that checks the test feeds against both schemas and runs Greenlight's EPIP rules.

### Changed
- The CLI uses `stream` for zip and folder input alike. `-location-system`, `-fare-frame`, `-general-frame`, `-streaming` and `-batch-size` are accepted and ignored.
- Frames carry EPIP's frame types (`epip:EU_PI_NETWORK_OFFER` and the rest).
- Each `ServiceJourneyPattern` names its line with `RouteView/LineRef` instead of `RouteRef`, as NAPCORE's "LineRef should be filled" asks.
- Feeds that break GTFS rules still give valid EPIP: bad UTF-8 and forbidden control characters become U+FFFD, station-level transfers expand to platforms, trip-to-trip interchange IDs include both stops, platforms with a missing station become their own `StopPlace`, and trips on unknown routes or stops are left out and counted.
- Extended route types (100 to 1400) map to their mode instead of `bus`.
- Untimed stops get times interpolated by distance, as Greenlight's `everyStopPointHaveArrivalAndDepartureTime` needs.
- Transfers: walk time is `WalkTransferDuration`; trip-to-trip transfers are `ServiceJourneyInterchange` with `StaySeated` or `StandardTransferTime`.
- Stops no trip serves get `StopPlaceType` `other`; routes with no trips are no longer written as `Line`s.
- IDs of routes, quays, journey patterns and route points change once.

### Deprecated
- `converter.Converter.Convert`: its output does not follow EPIP. Use `stream.ConvertZip` or `stream.ConvertFS`.

## [1.0.4] - 2026-09-11

### Fixed
- Authority, Operator, DataSource and ResponsibilitySet are now built from `agency.txt` (one Authority/Operator per agency) instead of a hardcoded "Sofia Transport" placeholder with fake contact details.
- Lines reference the Authority/Operator of their own `agency_id`; StopPlaces reference the primary agency's Authority instead of an empty ref.
- Frame IDs are scoped to the primary agency (e.g. `BG::CompositeFrame:BDZ::`) instead of `SOFIA`.
- Empty ContactDetails fields are omitted from the output.

## [1.0.0] - 2025-01-10

### Added

#### Core Functionality
- **GTFS to NeTEx Converter**: Complete conversion engine for transforming GTFS transit data to NeTEx format
- **High-Performance Architecture**: O(1) lookup indices using hash maps for optimal performance (390K+ entities/second)
- **Configurable Converter**: Support for any transit agency worldwide via configurable participant reference
- **Dynamic ID Generation**: Flexible ID generation system supporting multiple countries/agencies (BG, FR, US_NY, UK, JP)

#### CLI Application
- **Command-line Interface**: Production-ready CLI tool with comprehensive flags
- **Verbose Mode**: Detailed logging and statistics output
- **Report Generation**: JSON and text report formats with comprehensive statistics
- **Multiple Output Formats**: Support for XML output and JSON reports (text, json, both)

#### Conversion Features
- **Complete GTFS Support**:
  - `agency.txt` → Operator, Authority
  - `stops.txt` → StopPlace, Quay, ScheduledStopPoint
  - `routes.txt` → Line, Route
  - `trips.txt` → ServiceJourney
  - `stop_times.txt` → TimetabledPassingTime
  - `calendar_dates.txt` → DayType, OperatingPeriod
  - `transfers.txt` → Connection, InterchangeRule
  - `pathways.txt` → PathLink
  - `shapes.txt` → Loaded (not converted)
  - `fare_attributes.txt` → Tariff (with flag)

- **Missing Entity Generation**: Automatic generation of NeTEx-required entities not present in GTFS
- **Service Calendar Inference**: Smart inference of service patterns from calendar_dates.txt
- **Transport Mode Mapping**: Complete mapping of GTFS route types to NeTEx transport modes

#### NeTEx Structure
- **Frame Organization**: Proper NeTEx frame structure
  - ResourceFrame (operators, authorities)
  - ServiceCalendarFrame (day types, operating periods)
  - ServiceFrame (lines, routes, journey patterns)
  - SiteFrame (stop places, quays)
  - TimetableFrame (service journeys, passing times)

#### Reporting & Statistics
- **Comprehensive JSON Reports**: Detailed conversion statistics
  - Summary with status, duration, participant info
  - Input statistics (records by file type)
  - Output statistics (entities by type)
  - Performance metrics (throughput, duration)
  - Entity mapping (conversion success rates)
  - Warnings and errors

- **Performance Tracking**: Built-in performance monitoring
  - Records per second
  - Total duration
  - Memory usage tracking
  - Stage-wise timing

#### Testing
- **72.6% Test Coverage**: Comprehensive test suite
  - Unit tests for all core functions
  - Integration tests with real GTFS data
  - Performance benchmarks
  - Error handling tests
  - Multi-country configuration tests

- **Test Files**:
  - `converter_test.go` - Core converter unit tests
  - `report_test.go` - Report generation tests
  - `integration_test.go` - End-to-end integration tests
  - `lookup_indices_test.go` - Performance tests
  - `id_generator_test.go` - ID generation tests

#### Code Quality
- **Named Constants**: All magic numbers and strings replaced with named constants
- **Lookup Indices**: O(1) performance optimization for large datasets
- **Error Handling**: Comprehensive error checking and handling
- **Documentation**: Inline documentation and code comments
- **Linter Compliance**: Passes golangci-lint with all checks enabled

### Performance

- **Benchmark Results** (Sofia GTFS dataset):
  - Input: 1,593,965 records (10 files)
  - Output: 3,194,328 NeTEx entities
  - Processing time: 4.09 seconds
  - Throughput: 390,685 entities/second
  - Success rate: 100%

- **Optimization Techniques**:
  - O(1) lookup indices using hash maps
  - Pre-allocated slices for known sizes
  - Single-pass processing
  - Minimal memory allocations

### Project Structure

```
gtfs-netex-converter/
├── cmd/gtfs-netex-converter/  # CLI application
├── converter/                 # Conversion logic
│   ├── converter.go          # Main orchestration
│   ├── lookup_indices.go     # O(1) lookup system
│   ├── id_generator.go       # Dynamic ID generation
│   ├── report.go             # JSON reporting
│   ├── constants.go          # Named constants
│   ├── routes.go             # Route conversion
│   ├── stops.go              # Stop conversion
│   ├── calendar.go           # Calendar conversion
│   └── *_test.go             # Test files
├── gtfs/                     # GTFS type definitions
├── netex/                    # NeTEx type definitions
├── examples/                 # Usage examples
├── Makefile                  # Build automation
├── README.md                 # Documentation
└── CHANGELOG.md             # This file
```

### Configuration

- **Default Values**:
  - Participant: BG (Bulgaria)
  - Timezone: Europe/Sofia
  - Language: en
  - Location System: EPSG:4326

- **CLI Flags**:
  - `-input`: Input GTFS directory (required)
  - `-output`: Output NeTEx XML file (default: output.xml)
  - `-report`: JSON report file (optional)
  - `-report-format`: Report format - text, json, or both (default: text)
  - `-participant`: Participant reference code (default: BG)
  - `-timezone`: Default timezone (default: Europe/Sofia)
  - `-language`: Default language (default: en)
  - `-verbose`: Enable verbose output

### Dependencies

- Go 1.21 or higher
- Standard library only (no external dependencies for core functionality)

### Known Limitations

- **Shapes**: GTFS shapes.txt is loaded but not converted to NeTEx
- **Fares**: Fare conversion requires explicit flag and is not fully implemented
- **Missing GTFS Fields**: Some optional GTFS fields not mapped to NeTEx
  - `route_desc`, `route_url`, `route_color`, `route_text_color`, `route_sort_order`
  - `stop_desc`, `stop_url`, `zone_id`
  - `platform_code`, `wheelchair_boarding`

### Security

- MD5 hashing used only for non-cryptographic ID generation
- File permissions set to 0600 for output files
- Directory permissions set to 0750
- Input validation for all file operations

### Compatibility

- **Tested with GTFS feeds from**:
  - Sofia, Bulgaria (Urban Mobility Center)
  - Multiple transit agencies worldwide
  - Large datasets (1.6M+ records)

- **NeTEx Compliance**: Generates valid NeTEx XML according to DATA4PT specification

## [Unreleased]

### Planned Features

- Complete shapes.txt to NeTEx LinkSequence conversion
- Full fare frame generation with complex fare structures
- Support for GTFS frequencies.txt
- Calendar.txt support (in addition to calendar_dates.txt)
- Additional accessibility field mappings
- Route color and description preservation in KeyList
- Performance optimizations for datasets > 10M records
- Incremental conversion support for updates
- Validation against NeTEx XSD schema
- Multi-language support for entity names

### Known Issues

- None reported

---

## Release Notes

### v1.0.0 - Initial Production Release

This is the first production-ready release of the GTFS to NeTEx converter. The converter has been extensively tested with real-world data and is ready for use in production environments.

**Key Highlights**:
- ✅ Production-ready with 72.6% test coverage
- ✅ High performance (390K+ entities/second)
- ✅ Configurable for any transit agency worldwide
- ✅ Comprehensive JSON reporting
- ✅ Passes all linter checks (golangci-lint)
- ✅ Battle-tested with Sofia's 1.6M record dataset

**Getting Started**:
```bash
# Clone and build
git clone <repository-url>
cd gtfs-netex-converter
make build

# Convert GTFS to NeTEx
./bin/gtfs-netex-converter \
  -input ./gtfs_data \
  -output ./output.xml \
  -report ./report.json \
  -participant BG \
  -verbose
```

**Library Usage**:
```go
config := &converter.Config{
    ParticipantRef:  "BG",
    DefaultTimezone: "Europe/Sofia",
    DefaultLanguage: "en",
}

conv := converter.NewConverter(config)
netexData, err := conv.Convert("./gtfs_data")
report := conv.GenerateReport()
```

---

## Contributing

When contributing to this project, please:
1. Update this CHANGELOG.md with your changes
2. Follow the format: Added/Changed/Deprecated/Removed/Fixed/Security
3. Add entries under [Unreleased] section
4. Reference issue numbers where applicable

## Links

- [Repository](https://github.com/theoremus-urban-solutions/gtfs-netex-converter)
- [GTFS Specification](https://gtfs.org/reference/static/)
- [NeTEx Documentation](http://netex-cen.eu/)
- [DATA4PT GTFS-NeTEx Mapping](https://data4pt.org/)

[1.0.0]: https://github.com/theoremus-urban-solutions/gtfs-netex-converter/releases/tag/v1.0.0
[Unreleased]: https://github.com/theoremus-urban-solutions/gtfs-netex-converter/compare/v1.0.0...HEAD
