# GTFS to NeTEx Converter

A high-performance Go library and CLI tool for converting GTFS (General Transit Feed Specification) data to NeTEx (Network Timetable Exchange) format.

## What It Does

Transforms public transit data from GTFS format (used by Google Maps, transit apps) to NeTEx format (European standard for transit data exchange). Handles complete transit networks including stops, routes, schedules, and service calendars.

**Key Features:**
- 🚀 **High Performance**: O(1) lookups, processes 390K+ entities/second
- 🌍 **Configurable**: Works with any transit agency worldwide
- 📊 **Comprehensive Reporting**: JSON reports with statistics and performance metrics
- ✅ **Battle-Tested**: 72.6% test coverage, validated with real-world data
- 🔧 **Production-Ready**: Used for Sofia's 1.6M record transit dataset

## Quick Start

### Installation

```bash
# Clone and build
git clone <repository-url>
cd gtfs-netex-converter
make build

# Or use go install
go install ./cmd/gtfs-netex-converter
```

### Basic Usage

```bash
# Convert GTFS to NeTEx
./bin/gtfs-netex-converter \
  -input ./sofia_gtfs \
  -output ./sofia_netex.xml \
  -participant BG \
  -verbose

# With JSON report
./bin/gtfs-netex-converter \
  -input ./gtfs_data \
  -output ./output.xml \
  -report ./report.json \
  -report-format json \
  -participant BG
```

### As a Library

```go
package main

import (
    "gtfs-netex-converter/converter"
    "encoding/xml"
    "os"
)

func main() {
    // Configure converter
    config := &converter.Config{
        ParticipantRef:  "BG",
        DefaultTimezone: "Europe/Sofia",
        DefaultLanguage: "en",
    }

    // Convert
    conv := converter.NewConverter(config)
    netexData, err := conv.Convert("./gtfs_data")
    if err != nil {
        panic(err)
    }

    // Generate report
    report := conv.GenerateReport()
    jsonReport, _ := report.ToJSON()
    os.WriteFile("report.json", jsonReport, 0644)

    // Save NeTEx XML
    xmlData, _ := xml.MarshalIndent(netexData, "", "  ")
    os.WriteFile("output.xml", xmlData, 0644)
}
```

## How It Works

### 1. Load GTFS Data
Reads all GTFS files (stops, routes, trips, stop_times, calendar_dates, etc.) from input directory.

### 2. Build Lookup Indices
Creates O(1) hash map indices for fast entity lookups (critical for performance with large datasets).

### 3. Convert Entities
Transforms GTFS entities to NeTEx equivalents:

| GTFS | → | NeTEx |
|------|---|-------|
| Agency | → | Operator, Authority |
| Stop | → | StopPlace, Quay, ScheduledStopPoint |
| Route | → | Line, Route |
| Trip | → | ServiceJourney |
| StopTime | → | TimetabledPassingTime |
| CalendarDate | → | DayType, OperatingPeriod |

### 4. Generate NeTEx Structure
Organizes converted data into NeTEx frames:
- **ResourceFrame**: Organizations, operators
- **ServiceCalendarFrame**: Service calendars, day types
- **ServiceFrame**: Lines, routes, journey patterns
- **SiteFrame**: Stop places, quays
- **TimetableFrame**: Service journeys, passing times

### 5. Output XML + Report
Generates valid NeTEx XML and optional JSON report with statistics.

## CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `-input` | (required) | Input directory with GTFS files |
| `-output` | `output.xml` | Output NeTEx XML file |
| `-report` | | Optional JSON report file |
| `-report-format` | `text` | Report format: `text`, `json`, or `both` |
| `-participant` | `BG` | Participant reference (agency/country code) |
| `-timezone` | `Europe/Sofia` | Default timezone |
| `-language` | `en` | Default language code |
| `-verbose` | `false` | Enable verbose output |

## JSON Report Example

```json
{
  "summary": {
    "status": "success",
    "duration_seconds": 4.09,
    "participant_ref": "BG",
    "converter_version": "1.0.0"
  },
  "input_statistics": {
    "total_records": 1593965,
    "agencies": 1,
    "stops": 4315,
    "routes": 180,
    "trips": 25270,
    "stop_times": 588233
  },
  "output_statistics": {
    "authorities": 1,
    "stop_places": 4315,
    "lines": 180,
    "service_journeys": 25270
  },
  "performance": {
    "records_per_second": 390685.39
  },
  "entity_mapping": {
    "Agency->Authority": {
      "input_count": 1,
      "output_count": 1,
      "success_rate": 100
    }
  }
}
```

## Performance

Tested with Sofia transit dataset:
- **Input**: 1.6M GTFS records (10 files)
- **Processing Time**: 4.09 seconds
- **Throughput**: 390,685 entities/second
- **Success Rate**: 100%

Optimization techniques:
- O(1) lookup indices using hash maps
- Pre-allocated slices for known sizes
- Single-pass processing where possible
- Minimal memory allocations

## Testing

```bash
# Run all tests
make test

# Run with coverage
go test ./converter/... -cover

# Run specific test
go test ./converter/... -run TestEndToEndConversion -v
```

Test coverage: **72.6%**

Includes:
- Unit tests for all converters
- Integration tests with real GTFS data
- Performance benchmarks
- Error handling tests
- Multi-country configuration tests

## Project Structure

```
gtfs-netex-converter/
├── cmd/gtfs-netex-converter/  # CLI application
├── converter/                 # Conversion logic
│   ├── converter.go          # Main orchestration
│   ├── lookup_indices.go     # O(1) lookup system
│   ├── id_generator.go       # Dynamic ID generation
│   ├── report.go             # JSON reporting
│   ├── constants.go          # Named constants
│   └── *_test.go             # Test files
├── gtfs/                     # GTFS type definitions
├── netex/                    # NeTEx type definitions
└── examples/                 # Usage examples
```

## Supported GTFS Files

| File | Status | Notes |
|------|--------|-------|
| agency.txt | ✅ Required | Converted to Operator |
| stops.txt | ✅ Required | Converted to StopPlace/Quay |
| routes.txt | ✅ Required | Converted to Line |
| trips.txt | ✅ Required | Converted to ServiceJourney |
| stop_times.txt | ✅ Required | Converted to PassingTime |
| calendar_dates.txt | ✅ Supported | Converted to DayType |
| shapes.txt | ⚠️ Loaded | Not converted to NeTEx |
| transfers.txt | ✅ Supported | Converted to Connection |
| pathways.txt | ✅ Supported | Converted to PathLink |
| fare_attributes.txt | ⚠️ Loaded | Requires `-fare-frame` flag |

## Requirements

- Go 1.21 or higher
- Valid GTFS dataset with required files

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure `make test` passes
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Resources

- [GTFS Specification](https://gtfs.org/reference/static/)
- [NeTEx Documentation](http://netex-cen.eu/)
- [DATA4PT GTFS-NeTEx Mapping](https://data4pt.org/)

## Support

For questions or issues:
1. Check the [examples](./examples/) directory
2. Review test files for usage patterns
3. Open an issue on GitHub
