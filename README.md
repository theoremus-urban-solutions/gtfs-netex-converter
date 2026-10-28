# GTFS to NeTEx Converter

A high-performance Go library and CLI tool for converting GTFS (General Transit Feed Specification) data to NeTEx (Network Timetable Exchange) format.

## What It Does

Transforms public transit data from GTFS format (used by Google Maps, transit apps) to NeTEx format (European standard for transit data exchange). Handles complete transit networks including stops, routes, schedules, and service calendars.

**Key Features:**
- ✅ **EPIP output**: passes the strict EPIP schema and the NeTEx v2.0.0 schema, checked in CI
- 🚀 **Streaming**: Sofia's 1.6M-record feed converts in about 5 s with under 200 MB of memory
- 📅 **Exact calendars**: every service day survives the conversion, checked day by day in the tests
- 📊 **Reporting**: JSON report with counts, timings, and every piece of GTFS left out

## NeTEx Profile: EPIP

The converter's output follows the **European Passenger Information Profile (EPIP)**, CEN/TS 16614-4. EPIP is the European minimum NeTEx profile for passenger information: stops, lines, routes, timetables and calendars.

**Why EPIP.** Delegated Regulation (EU) 2024/490, Article 4(2), says NeTEx data on a national access point "shall be represented through minimum EU profiles or national profiles". Bulgaria has no national NeTEx profile, so the EU profile applies.

**Schemas.** Every output must pass both checks with 0 errors:

| Check | Schema |
|-------|--------|
| Profile | `NeTEx_publication_EPIP.xsd` from [NeTEx-Profile-EPIP](https://github.com/TransmodelEcosystem/NeTEx-Profile-EPIP). It is built on NeTEx 1.3.1, and the repository is archived. |
| Base | `xsd/NeTEx_publication.xsd` from [NeTEx v2.0.0](https://github.com/TransmodelEcosystem/NeTEx/releases/tag/v2.0.0) |

EPIP is a subset of NeTEx, so a file that passes EPIP also passes the full v2.0 schema.

```bash
xmllint --noout --schema NeTEx-Profile-EPIP/NeTEx_publication_EPIP.xsd output.xml
xmllint --noout --schema NeTEx/xsd/NeTEx_publication.xsd output.xml
# For files of a few hundred MB, add --stream to keep memory low.
```

**Field mapping.** GTFS fields map to NeTEx as the [DATA4PT GTFS-NeTEx mapping](./2024-June_DATA4PT_GTFS-NeTEx-Mapping_vf.pdf) describes. Where DATA4PT uses an element EPIP does not allow, EPIP wins.

**Status.** Since v1.1.0 the output conforms. CI converts four test feeds, checks each against both schemas, and runs Greenlight's 13 EPIP rules on the 8-route Sofia sample. The full Sofia feed also passes both schemas. What EPIP cannot carry, such as blocks, pathways and fares, is left out and counted in the report; see [UNMAPPABLE_FEATURES.md](UNMAPPABLE_FEATURES.md).

The `converter` package's `Convert` is the pre-1.1.0 converter. Its output does not conform and it is deprecated; use `stream.ConvertZip`.

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
# Convert a GTFS zip (or a folder of GTFS files) to EPIP NeTEx
./bin/gtfs-netex-converter \
  -input ./sofia_gtfs.zip \
  -output ./sofia_netex.xml \
  -participant BG \
  -verbose

# Zipped output with a JSON report
./bin/gtfs-netex-converter \
  -input ./gtfs.zip \
  -output ./netex.zip \
  -report ./report.json \
  -report-format json
```

### As a Library

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/theoremus-urban-solutions/gtfs-netex-converter/stream"
)

func main() {
    out, err := os.Create("output.xml")
    if err != nil {
        log.Fatal(err)
    }
    defer out.Close()

    // stream.ConvertZip takes an io.ReaderAt and its size; stream.ConvertFS takes a folder as fs.FS.
    report, err := stream.ConvertZipFile(context.Background(), "gtfs.zip", out, stream.Config{
        Codespace: "BG",
        TimeZone:  "Europe/Sofia",
        Language:  "bg",
    })
    if err != nil {
        log.Fatal(err)
    }
    jsonReport, _ := report.ToJSON()
    os.WriteFile("report.json", jsonReport, 0o644)
}
```

## How It Works

`stream` reads the GTFS files one at a time and keeps only compact tables in memory. Stop times are packed into integers and shapes into float slices. It then writes the XML straight to the destination, frame by frame, without building a document tree. The frames, and where each GTFS field goes, are in [GTFS-NeTEx-Mapping-Reference.md](GTFS-NeTEx-Mapping-Reference.md).

## CLI Options

| Flag | Default | Description |
|------|---------|-------------|
| `-input` | (required) | GTFS zip, or a folder of GTFS files |
| `-output` | `output.xml` | Output NeTEx XML file; a name ending in `.zip` writes a zip holding one XML file |
| `-report` | | Optional JSON report file |
| `-report-format` | `text` | Report format: `text`, `json`, or `both` |
| `-participant` | `BG` | Codespace, ID prefix and participant reference |
| `-timezone` | `Europe/Sofia` | Default timezone |
| `-language` | `bg` | Default language code |
| `-publication-time` | now | `PublicationTimestamp` as `2006-01-02T15:04:05`; fix it to get the same bytes on every run |
| `-verbose` | `false` | Enable verbose output |

`-location-system`, `-fare-frame`, `-general-frame`, `-streaming` and `-batch-size` are still accepted and ignored.

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

Full Sofia feed (1.6M GTFS records, 29,156 trips), measured on 2026-10-28 with `/usr/bin/time -v`:

| Output | Time | Peak memory | Size |
|--------|------|-------------|------|
| XML | 5.3 s | 194 MB | 339 MB |
| zip | 8.1 s | 197 MB | 20 MB |

The pre-1.1.0 `Convert` needed about 1.1 GB for the same feed.

## Testing

```bash
make test
go test ./stream/... -cover
# After a deliberate output change, rewrite the golden files and review the diff:
go test ./stream/... -run TestGolden -update
```

`stream` has 93.6% test coverage. Its tests include:
- golden output for three small feeds
- a calendar check that expands every service back into dates, through the weekdays and through `ValidDayBits` alone, and compares them with GTFS
- shape cutting on a straight line, a loop, and a stop far off the shape
- an onward `ServiceLink` for every stop but the last

## Project Structure

```
gtfs-netex-converter/
├── cmd/gtfs-netex-converter/  # CLI application
├── stream/                    # EPIP converter (ConvertZip, ConvertFS)
│   └── testdata/             # Test feeds and golden NeTEx
├── converter/                 # Deprecated tree converter and the report types
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
| agency.txt | ✅ Converted | Authority and Operator |
| stops.txt | ✅ Required | StopPlace, Quay, ScheduledStopPoint |
| routes.txt | ✅ Required | Line, Route |
| trips.txt | ✅ Required | ServiceJourney |
| stop_times.txt | ✅ Required | ServiceJourneyPattern, TimetabledPassingTime |
| calendar.txt, calendar_dates.txt | ✅ Converted | DayType, UicOperatingPeriod, DayTypeAssignment |
| shapes.txt | ✅ Converted | ServiceLink geometry, cut at stops |
| transfers.txt | ✅ Converted | Connection, ServiceJourneyInterchange |
| feed_info.txt | ✅ Converted | validity period |
| frequencies.txt, translations.txt, levels.txt | ⚠️ Not yet | planned for v1.2.0, counted in the report |
| pathways.txt, fare_attributes.txt, fare_rules.txt | ❌ Left out | no place in EPIP, counted in the report |

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
- [NeTEx-Profile-EPIP (EPIP schema)](https://github.com/TransmodelEcosystem/NeTEx-Profile-EPIP)
- [NeTEx v2.0.0](https://github.com/TransmodelEcosystem/NeTEx/releases/tag/v2.0.0)
- [Delegated Regulation (EU) 2024/490](https://eur-lex.europa.eu/eli/reg_del/2024/490/oj/eng)

## Support

For questions or issues:
1. Check the [examples](./examples/) directory
2. Review test files for usage patterns
3. Open an issue on GitHub
