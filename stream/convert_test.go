package stream

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "rewrite the golden NeTEx files in testdata")

var testCfg = Config{Codespace: "BG", TimeZone: "Europe/Sofia", Language: "bg", PublicationTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

func convertFixture(t *testing.T, name string) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := ConvertZipFile(context.Background(), filepath.Join("testdata", name+".zip"), &buf, testCfg); err != nil {
		t.Fatalf("convert %s: %v", name, err)
	}
	return buf.Bytes()
}

func TestGolden(t *testing.T) {
	for _, name := range []string{"allfields", "calendar", "sofia-example"} {
		t.Run(name, func(t *testing.T) {
			got := convertFixture(t, name)
			path := filepath.Join("testdata", name+".xml")
			if *update {
				if err := os.WriteFile(path, got, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			want, err := os.ReadFile(path) //nolint:gosec // test fixture
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("%s differs from %s; run go test ./stream -update after checking the change", name, path)
			}
		})
	}
}

// TestCalendarDays expands every DayType back into dates, once through weekdays and once through ValidDayBits alone, and compares both with GTFS.
func TestCalendarDays(t *testing.T) {
	for _, name := range []string{"allfields", "calendar", "sofia-example", "sofia-slice"} {
		t.Run(name, func(t *testing.T) {
			want := gtfsDays(t, filepath.Join("testdata", name+".zip"))
			got, fromBits := netexDays(t, convertFixture(t, name))
			for sid, days := range want {
				if !sameDays(days, got[sid]) {
					t.Errorf("service %s: NeTEx gives %d days, GTFS %d", sid, len(got[sid]), len(days))
				}
				if b, ok := fromBits[sid]; ok {
					in := map[int]bool{}
					for d := range days {
						if d >= b.lo && d <= b.hi {
							in[d] = true
						}
					}
					if !sameDays(in, b.days) {
						t.Errorf("service %s: ValidDayBits give %d days, GTFS %d in the period", sid, len(b.days), len(in))
					}
				}
			}
		})
	}
}

// TestJourneyPatternLinks checks every stop but the last points to its onward ServiceLink, and every link exists.
func TestJourneyPatternLinks(t *testing.T) {
	d := xml.NewDecoder(bytes.NewReader(convertFixture(t, "sofia-slice")))
	links := map[string]bool{}
	var onward []string
	stops, refs, withGeom := 0, 0, 0
	inLink := ""
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch el := tok.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "ServiceLink":
				inLink = attr(el, "id")
				links[inLink] = true
			case "LineString":
				if inLink != "" {
					withGeom++
				}
			case "ServiceJourneyPattern":
				if stops > 0 && refs != stops-1 {
					t.Fatalf("pattern with %d stops has %d onward links", stops, refs)
				}
				stops, refs = 0, 0
			case "StopPointInJourneyPattern":
				stops++
			case "OnwardServiceLinkRef":
				refs++
				onward = append(onward, attr(el, "ref"))
			}
		case xml.EndElement:
			if el.Name.Local == "ServiceLink" {
				inLink = ""
			}
		}
	}
	for _, r := range onward {
		if !links[r] {
			t.Fatalf("OnwardServiceLinkRef %s has no ServiceLink", r)
		}
	}
	if len(links) == 0 || withGeom != len(links) {
		t.Fatalf("%d of %d ServiceLinks carry geometry", withGeom, len(links))
	}
}

func TestCutShape(t *testing.T) {
	// A straight shape north along one meridian, with three stops on it.
	shape := []float64{42.0, 23.0, 42.001, 23.0, 42.002, 23.0, 42.003, 23.0}
	pieces := cutShape(shape, [][2]float64{{42.0, 23.0}, {42.0015, 23.0}, {42.003, 23.0}})
	if len(pieces) != 2 {
		t.Fatalf("got %d pieces, want 2", len(pieces))
	}
	if got := pieces[0]; len(got) != 6 || got[4] != 42.0015 {
		t.Fatalf("first piece %v should end at the middle stop", got)
	}
	// A loop that starts and ends at the same stop: the last stop must take the end of the shape.
	loop := []float64{42.0, 23.0, 42.001, 23.0, 42.001, 23.001, 42.0, 23.001, 42.0, 23.0}
	pieces = cutShape(loop, [][2]float64{{42.0, 23.0}, {42.001, 23.0005}, {42.0, 23.0}})
	if len(pieces) != 2 || len(pieces[1]) < 6 {
		t.Fatalf("loop pieces %v should run round the loop", pieces)
	}
	// A stop far from the shape gives no geometry.
	if cutShape(shape, [][2]float64{{42.0, 23.0}, {42.5, 23.5}}) != nil {
		t.Fatal("a stop 50 km off the shape should not cut it")
	}
}

func TestReportCountsLeftOut(t *testing.T) {
	rep, err := ConvertZipFile(context.Background(), filepath.Join("testdata", "allfields.zip"), io.Discard, testCfg)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, w := range rep.Warnings {
		got[w.Type] = w.Count
	}
	for _, k := range []string{"block_left_out", "transfer_not_possible_left_out", "guaranteed_left_out", "route_transfer_left_out", "not_converted_pathways"} {
		if got[k] == 0 {
			t.Errorf("report has no %s warning: %v", k, got)
		}
	}
}

func TestCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := ConvertZipFile(ctx, filepath.Join("testdata", "sofia-example.zip"), io.Discard, testCfg); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestMissingStopTimes(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range map[string]string{"routes.txt": "route_id\nR\n", "stops.txt": "stop_id\nS\n", "trips.txt": "trip_id,route_id,service_id\nT,R,X\n"} {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	_ = zw.Close()
	_, err := ConvertZip(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()), io.Discard, testCfg)
	if err == nil || !strings.Contains(err.Error(), "stop_times.txt") {
		t.Fatalf("got %v, want an error naming stop_times.txt", err)
	}
}

func attr(el xml.StartElement, name string) string {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func sameDays(a, b map[int]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for d := range a {
		if !b[d] {
			return false
		}
	}
	return true
}

// gtfsDays reads the service days straight from calendar.txt and calendar_dates.txt.
func gtfsDays(t *testing.T, path string) map[string]map[int]bool {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = zr.Close() }()
	read := func(name string) []map[string]string {
		f, err := zr.Open(name)
		if err != nil {
			return nil
		}
		defer func() { _ = f.Close() }()
		recs, err := csv.NewReader(f).ReadAll()
		if err != nil {
			t.Fatal(err)
		}
		var out []map[string]string
		for _, r := range recs[1:] {
			m := map[string]string{}
			for i, h := range recs[0] {
				m[strings.TrimPrefix(h, "\ufeff")] = r[i]
			}
			out = append(out, m)
		}
		return out
	}
	days := map[string]map[int]bool{}
	get := func(s string) map[int]bool {
		if days[s] == nil {
			days[s] = map[int]bool{}
		}
		return days[s]
	}
	names := []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}
	for _, r := range read("calendar.txt") {
		a, _ := dayNum(r["start_date"])
		b, _ := dayNum(r["end_date"])
		m := get(r["service_id"])
		for d := a; d <= b; d++ {
			if r[names[weekday(d)]] == "1" {
				m[d] = true
			}
		}
	}
	for _, r := range read("calendar_dates.txt") {
		d, _ := dayNum(r["date"])
		m := get(r["service_id"])
		switch r["exception_type"] {
		case "1":
			m[d] = true
		case "2":
			delete(m, d)
		}
	}
	return days
}

type bitDays struct {
	lo, hi int
	days   map[int]bool
}

// netexDays expands the ServiceCalendarFrame into service days per GTFS service_id.
func netexDays(t *testing.T, doc []byte) (map[string]map[int]bool, map[string]bitDays) {
	t.Helper()
	var cal struct {
		DayTypes []struct {
			ID   string `xml:"id,attr"`
			Days string `xml:"properties>PropertyOfDay>DaysOfWeek"`
		} `xml:"ServiceCalendar>dayTypes>DayType"`
		Periods []struct {
			ID   string `xml:"id,attr"`
			From string `xml:"FromDate"`
			To   string `xml:"ToDate"`
			Bits string `xml:"ValidDayBits"`
		} `xml:"ServiceCalendar>operatingPeriods>UicOperatingPeriod"`
		Assignments []struct {
			Period struct {
				Ref string `xml:"ref,attr"`
			} `xml:"OperatingPeriodRef"`
			Date    string `xml:"Date"`
			DayType struct {
				Ref string `xml:"ref,attr"`
			} `xml:"DayTypeRef"`
			Available bool `xml:"isAvailable"`
		} `xml:"ServiceCalendar>dayTypeAssignments>DayTypeAssignment"`
	}
	d := xml.NewDecoder(bytes.NewReader(doc))
	for {
		tok, err := d.Token()
		if err != nil {
			t.Fatalf("no ServiceCalendarFrame: %v", err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "ServiceCalendarFrame" {
			if err := d.DecodeElement(&cal, &se); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	sid := func(ref string) string { return strings.TrimSuffix(strings.SplitN(ref, "day_type_", 2)[1], "::") }
	weekdaysOf := map[string]map[string]bool{}
	for _, dt := range cal.DayTypes {
		weekdaysOf[dt.ID] = map[string]bool{}
		for _, w := range strings.Fields(dt.Days) {
			weekdaysOf[dt.ID][w] = true
		}
	}
	type period struct {
		lo, hi int
		bits   string
	}
	periods := map[string]period{}
	for _, p := range cal.Periods {
		lo, _ := dayNum(strings.ReplaceAll(p.From[:10], "-", ""))
		hi, _ := dayNum(strings.ReplaceAll(p.To[:10], "-", ""))
		periods[p.ID] = period{lo, hi, p.Bits}
	}
	got := map[string]map[int]bool{}
	fromBits := map[string]bitDays{}
	for _, a := range cal.Assignments {
		s := sid(a.DayType.Ref)
		if got[s] == nil {
			got[s] = map[int]bool{}
		}
		if a.Period.Ref != "" {
			p := periods[a.Period.Ref]
			if len(p.bits) != p.hi-p.lo+1 {
				t.Fatalf("%s: %d bits for %d days", a.Period.Ref, len(p.bits), p.hi-p.lo+1)
			}
			b := bitDays{p.lo, p.hi, map[int]bool{}}
			for d := p.lo; d <= p.hi; d++ {
				on := p.bits[d-p.lo] == '1'
				if on {
					b.days[d] = true
				}
				if on && weekdaysOf[a.DayType.Ref][weekdayNames[weekday(d)]] {
					setDay(got[s], d, a.Available)
				}
			}
			fromBits[s] = b
			continue
		}
		dn, _ := dayNum(strings.ReplaceAll(a.Date, "-", ""))
		setDay(got[s], dn, a.Available)
	}
	return got, fromBits
}

func setDay(m map[int]bool, d int, on bool) {
	if on {
		m[d] = true
	} else {
		delete(m, d)
	}
}

// TestEdgeFeeds converts small feeds that each break one rule real feeds break, and checks the XML is well-formed, every versioned ref resolves and the case is handled.
func TestEdgeFeeds(t *testing.T) {
	checks := map[string]func(doc string, warn map[string]int) string{
		"bad-chars": func(doc string, _ map[string]int) string {
			return want(strings.Contains(doc, "� cp1251 � tab"), "bad bytes are not replaced with U+FFFD")
		},
		"station-transfer": func(doc string, _ map[string]int) string {
			return want(strings.Count(doc, `scheduled_stop_point_S4::" version="1"></ScheduledStopPointRef></To>`) == 2, "station transfer is not expanded to both platforms")
		},
		"dup-interchange": func(doc string, _ map[string]int) string {
			return want(strings.Contains(doc, ":T1_T2_P2_P1::") && strings.Contains(doc, ":T1_T2_P2_P2::"), "interchanges sharing trips and from-stop are not both written")
		},
		"orphan-parent": func(doc string, warn map[string]int) string {
			return want(strings.Contains(doc, "stop_place_P2::") && warn["parent_station_not_a_station"] == 1, "platform with a missing station is not its own stop")
		},
		"trip-unknown-route": func(doc string, warn map[string]int) string {
			return want(!strings.Contains(doc, "TRIP_T4") && warn["trip_unknown_route_left_out"] == 1, "trip on an unknown route is not left out")
		},
		"transfer-unknown-trip": func(doc string, warn map[string]int) string {
			return want(!strings.Contains(doc, "TRIP_T9") && warn["interchange_unknown_trip_left_out"] == 1, "transfer to an unknown trip is not left out")
		},
		"untimed-middle": func(doc string, warn map[string]int) string {
			return want(strings.Contains(doc, `TPT_T1_1::" version="1"><StopPointInJourneyPatternRef`) && strings.Contains(doc, "</StopPointInJourneyPatternRef><ArrivalTime>08:10:03</ArrivalTime><DepartureTime>08:10:03</DepartureTime>") &&
				warn["times_interpolated"] == 1, "untimed stop is not interpolated by distance")
		},
		"rail-extended": func(doc string, _ map[string]int) string {
			return want(strings.Contains(doc, "<Name>Rail</Name>") && !strings.Contains(doc, "vehicle_type_100::\" version=\"1\"><Name>Bus"), "route_type 100 is not rail")
		},
	}
	files, err := filepath.Glob(filepath.Join("testdata", "edge", "*.zip"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no edge feeds: %v", err)
	}
	for _, path := range files {
		name := strings.TrimSuffix(filepath.Base(path), ".zip")
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			rep, err := ConvertZipFile(context.Background(), path, &buf, testCfg)
			if err != nil {
				t.Fatalf("convert: %v", err)
			}
			if msg := danglingRefs(buf.Bytes()); msg != "" {
				t.Fatal(msg)
			}
			warn := map[string]int{}
			for _, w := range rep.Warnings {
				warn[w.Type] = w.Count
			}
			if check := checks[name]; check != nil {
				if msg := check(buf.String(), warn); msg != "" {
					t.Fatal(msg)
				}
			}
		})
	}
}

func want(ok bool, msg string) string {
	if ok {
		return ""
	}
	return msg
}

// danglingRefs parses the document, which fails on bad XML, and returns the first ref with a version that names no id, as the schema's keyrefs would.
func danglingRefs(doc []byte) string {
	d := xml.NewDecoder(bytes.NewReader(doc))
	ids := map[string]bool{}
	var refs []string
	for {
		tok, err := d.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "not well-formed XML: " + err.Error()
		}
		if el, ok := tok.(xml.StartElement); ok {
			if id := attr(el, "id"); id != "" {
				ids[id] = true
			}
			if r := attr(el, "ref"); r != "" && attr(el, "version") != "" {
				refs = append(refs, el.Name.Local+" "+r)
			}
		}
	}
	for _, r := range refs {
		if !ids[r[strings.IndexByte(r, ' ')+1:]] {
			return "dangling " + r
		}
	}
	return ""
}
