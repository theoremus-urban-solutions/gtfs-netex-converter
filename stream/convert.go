// Package stream converts a GTFS zip into one NeTEx document that follows the European Passenger Information Profile (EPIP).
// It writes the XML while it reads, without building an object tree.
package stream

import (
	"archive/zip"
	"context"
	"crypto/md5" //nolint:gosec // non-cryptographic IDs
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/converter"
)

// Version is written into the report.
const Version = "1.1.0"

// Config holds the values a GTFS feed does not carry.
type Config struct {
	Codespace       string    // ID prefix, codespace and ParticipantRef; "BG" when empty
	TimeZone        string    // "Europe/Sofia" when empty
	Language        string    // "bg" when empty
	PublicationTime time.Time // time.Now() when zero
}

// ConvertZipFile converts the GTFS zip at path; see ConvertZip.
func ConvertZipFile(ctx context.Context, path string, dst io.Writer, cfg Config) (*converter.ConversionReport, error) {
	f, err := os.Open(path) //nolint:gosec // caller chooses the input
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return ConvertZip(ctx, f, st.Size(), dst, cfg)
}

// ConvertZip reads a GTFS zip and writes one EPIP NeTEx XML document to dst.
func ConvertZip(ctx context.Context, src io.ReaderAt, size int64, dst io.Writer, cfg Config) (*converter.ConversionReport, error) {
	zr, err := zip.NewReader(src, size)
	if err != nil {
		return nil, fmt.Errorf("open GTFS zip: %w", err)
	}
	return ConvertFS(ctx, zr, dst, cfg)
}

// ConvertFS reads GTFS files from fsys, such as os.DirFS of an unzipped feed, and writes one EPIP NeTEx XML document to dst.
func ConvertFS(ctx context.Context, fsys fs.FS, dst io.Writer, cfg Config) (*converter.ConversionReport, error) {
	if cfg.Codespace == "" {
		cfg.Codespace = "BG"
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = "Europe/Sofia"
	}
	if cfg.Language == "" {
		cfg.Language = "bg"
	}
	if cfg.PublicationTime.IsZero() {
		cfg.PublicationTime = time.Now()
	}
	c := &conv{ctx: ctx, cfg: cfg, cs: cfg.Codespace, fs: fsys, o: newOut(dst), start: time.Now(),
		recs: map[string]int{}, warn: map[string]*converter.Warning{}, made: map[string]int{}}
	if err := c.run(); err != nil {
		return nil, err
	}
	return c.report(), nil
}

type tinfo struct {
	row      []string
	tid      uint32
	key, dir string
	shape    string // shape the pattern was cut with, empty when it has no geometry
}

type pattern struct {
	key   string
	first int // index into tinfos
	trip  int // trip index in span
	links []string
}

type org struct{ key, auth, oper string }

type link struct {
	id, from, to string
	geom         []float64
}

type conv struct {
	ctx   context.Context
	cfg   Config
	cs    string
	fs    fs.FS
	o     *out
	start time.Time

	agency, routes, stops, trips, calendar, feedInfo, transfers *table

	rows                 []stRow
	span                 [][2]int
	tripIx, stopIx, hsIx *interner
	tinfos               []tinfo
	patterns             []pattern
	shapePts             map[string][]float64
	links                []link
	cal                  []calService
	calLo, calHi         int
	routeByID            map[string][]string
	stopByID             map[string][]string
	children             map[string][][]string // platforms of each station
	orgs                 []org
	orgOf                map[string]org

	recs map[string]int
	warn map[string]*converter.Warning
	made map[string]int
}

func (c *conv) id(typ, local string) string { return c.cs + "::" + typ + ":" + local + "::" }

func (c *conv) ssp(stop string) string {
	return c.id("ScheduledStopPoint", "scheduled_stop_point_"+stop)
}

// locType returns a stop's location_type, or -1 when stops.txt has no such stop.
func (c *conv) locType(id string) int {
	r, ok := c.stopByID[id]
	if !ok {
		return -1
	}
	return atoiDef(c.stops.get(r, "location_type"), 0)
}

// lone reports whether a stops.txt row is a platform with no station holding it.
func (c *conv) lone(r []string) bool {
	return atoiDef(c.stops.get(r, "location_type"), 0) == 0 && c.locType(c.stops.get(r, "parent_station")) != 1
}

func (c *conv) note(typ, msg string, n int) {
	if n <= 0 {
		return
	}
	if w := c.warn[typ]; w != nil {
		w.Count += n
		return
	}
	c.warn[typ] = &converter.Warning{Type: typ, Message: msg, Count: n}
}

func (c *conv) alive() error { return c.ctx.Err() }

func md5hex(s string) string { return fmt.Sprintf("%x", md5.Sum([]byte(s))) } //nolint:gosec

func gmlID(s string) string {
	b := []byte("LS_")
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == '.' {
			b = append(b, ch)
		} else {
			b = append(b, '_')
		}
	}
	return string(b)
}

var modeOfType = map[int]string{0: "tram", 1: "metro", 2: "rail", 3: "bus", 4: "water", 5: "cableway", 6: "cableway", 7: "funicular", 11: "trolleyBus", 12: "rail"}
var vehicleTypeName = map[int]string{0: "Tram", 1: "Metro", 2: "Rail", 3: "Bus", 4: "Ferry", 5: "Cable Car", 6: "Gondola", 7: "Funicular", 11: "Trolleybus", 12: "Monorail"}
var stopPlaceTypeOfMode = map[string]string{"bus": "onstreetBus", "trolleyBus": "onstreetBus", "tram": "onstreetTram", "metro": "metroStation",
	"rail": "railStation", "water": "ferryStop", "cableway": "liftStation", "funicular": "liftStation", "coach": "coachStation", "air": "airport"}
var stationTypeOfMode = map[string]string{"bus": "busStation", "trolleyBus": "busStation", "tram": "tramStation", "metro": "metroStation",
	"rail": "railStation", "water": "ferryPort", "cableway": "liftStation", "funicular": "liftStation", "coach": "coachStation", "air": "airport"}

// extendedModes maps each hundred of the extended route types (100 railway ... 1400 funicular) to its EPIP mode.
var extendedModes = map[int]string{1: "rail", 2: "coach", 3: "rail", 4: "metro", 5: "metro", 6: "metro", 7: "bus", 8: "trolleyBus",
	9: "tram", 10: "water", 11: "air", 12: "water", 13: "cableway", 14: "funicular"}
var vehicleNameOfMode = map[string]string{"rail": "Rail", "coach": "Coach", "metro": "Metro", "bus": "Bus", "trolleyBus": "Trolleybus",
	"tram": "Tram", "water": "Ferry", "air": "Aircraft", "cableway": "Cable Car", "funicular": "Funicular"}

// modeOK returns the EPIP mode of a route type, and false when the type is unknown and "bus" is a guess.
func modeOK(rt int) (string, bool) {
	if m, ok := modeOfType[rt]; ok {
		return m, true
	}
	if m, ok := extendedModes[rt/100]; ok && rt >= 100 {
		return m, true
	}
	return "bus", false
}

func mode(rt int) string {
	m, _ := modeOK(rt)
	return m
}

func vehicleName(rt int) string {
	if n, ok := vehicleTypeName[rt]; ok {
		return n
	}
	return vehicleNameOfMode[mode(rt)]
}

func agencyKey(id, name string, pos int) string {
	raw := id
	if raw == "" {
		raw = name
	}
	if raw == "" {
		raw = fmt.Sprintf("UNKNOWN_%d", pos)
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func hms(t int32) (string, int) {
	if t < 0 {
		return "", 0
	}
	d := int(t) / 86400
	t %= 86400
	return fmt.Sprintf("%02d:%02d:%02d", t/3600, t/60%60, t%60), d
}

func (c *conv) run() error {
	if err := c.load(); err != nil {
		return err
	}
	if err := c.alive(); err != nil {
		return err
	}
	c.header()
	c.resourceFrame()
	c.calendarFrame()
	c.serviceFrame()
	c.siteFrame()
	if err := c.timetableFrame(); err != nil {
		return err
	}
	c.o.s("</frames></CompositeFrame></dataObjects></PublicationDelivery>")
	if err := c.o.flush(); err != nil {
		return fmt.Errorf("write NeTEx: %w", err)
	}
	return nil
}

func (c *conv) load() error {
	var err error
	for _, t := range []struct {
		dst  **table
		name string
	}{{&c.agency, "agency.txt"}, {&c.routes, "routes.txt"}, {&c.stops, "stops.txt"}, {&c.trips, "trips.txt"},
		{&c.calendar, "calendar.txt"}, {&c.feedInfo, "feed_info.txt"}, {&c.transfers, "transfers.txt"}} {
		if *t.dst, err = readTable(c.fs, t.name); err != nil {
			return err
		}
		if n := len((*t.dst).rows); n > 0 || len((*t.dst).col) > 0 {
			c.recs[t.name] = n
		}
	}
	for _, need := range []string{"routes.txt", "stops.txt", "trips.txt"} {
		if _, ok := c.recs[need]; !ok {
			return fmt.Errorf("GTFS feed has no %s", need)
		}
	}
	if err := c.loadStopTimes(); err != nil {
		return err
	}
	c.routeByID = map[string][]string{}
	for _, r := range c.routes.rows {
		c.routeByID[c.routes.get(r, "route_id")] = r
	}
	c.stopByID = map[string][]string{}
	for _, r := range c.stops.rows {
		c.stopByID[c.stops.get(r, "stop_id")] = r
	}
	c.children = map[string][][]string{}
	orphans := 0
	for _, r := range c.stops.rows {
		if p := c.stops.get(r, "parent_station"); p != "" && c.locType(c.stops.get(r, "stop_id")) == 0 {
			if c.locType(p) == 1 {
				c.children[p] = append(c.children[p], r)
			} else {
				orphans++
			}
		}
	}
	c.note("parent_station_not_a_station", "Platforms whose parent_station is missing or not a station become stops of their own.", orphans)
	c.fillTimes()
	c.buildPatterns()
	if err := c.loadShapes(); err != nil {
		return err
	}
	c.buildLinks()
	if err := c.loadCalendar(); err != nil {
		return err
	}
	return c.countLeftOut()
}

func (c *conv) loadStopTimes() error {
	c.tripIx, c.stopIx, c.hsIx = newInterner(), newInterner(), newInterner()
	c.hsIx.id("")
	for _, r := range c.trips.rows {
		c.tripIx.id(c.trips.get(r, "trip_id"))
	}
	f, err := openCSV(c.fs, "stop_times.txt")
	if err != nil {
		return err
	}
	if f == nil {
		return errors.New("GTFS feed has no stop_times.txt")
	}
	defer f.close()
	c.rows = make([]stRow, 0, 1<<16)
	for {
		rec, err := f.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		c.rows = append(c.rows, stRow{
			trip: c.tripIx.id(f.get(rec, "trip_id")), stop: c.stopIx.id(f.get(rec, "stop_id")),
			seq: uint32(atoiDef(f.get(rec, "stop_sequence"), 0)), //nolint:gosec // GTFS sequences are non-negative
			arr: parseTime(f.get(rec, "arrival_time")), dep: parseTime(f.get(rec, "departure_time")),
			pickup: uint8(atoiDef(f.get(rec, "pickup_type"), 0)), drop: uint8(atoiDef(f.get(rec, "drop_off_type"), 0)), //nolint:gosec // GTFS enums are 0-3
			headsign: c.hsIx.id(f.get(rec, "stop_headsign")),
		})
	}
	c.recs["stop_times.txt"] = len(c.rows)
	slices.SortStableFunc(c.rows, func(a, b stRow) int {
		if a.trip != b.trip {
			return int(a.trip) - int(b.trip)
		}
		return int(a.seq) - int(b.seq)
	})
	c.span = make([][2]int, len(c.tripIx.strs))
	for i := range c.span {
		c.span[i] = [2]int{-1, -1}
	}
	for i := 0; i < len(c.rows); {
		j := i
		for j < len(c.rows) && c.rows[j].trip == c.rows[i].trip {
			j++
		}
		c.span[c.rows[i].trip] = [2]int{i, j}
		i = j
	}
	return nil
}

// buildPatterns groups trips by route, direction, stops, shape, headsigns and pickup and drop-off rules.
func (c *conv) buildPatterns() {
	patIdx := map[string]int{}
	var sb strings.Builder
	skipped, noRoute, noStop, noTimes := 0, 0, 0, 0
	platform := make([]bool, len(c.stopIx.strs))
	for i, id := range c.stopIx.strs {
		platform[i] = c.locType(id) == 0
	}
trips:
	for _, tr := range c.trips.rows {
		tid := c.tripIx.ids[c.trips.get(tr, "trip_id")]
		sp := c.span[tid]
		if c.trips.get(tr, "service_id") == "" || sp[0] < 0 {
			skipped++
			continue
		}
		if _, ok := c.routeByID[c.trips.get(tr, "route_id")]; !ok {
			noRoute++
			continue
		}
		if c.rows[sp[0]].dep < 0 || c.rows[sp[1]-1].arr < 0 {
			noTimes++
			continue
		}
		for j := sp[0]; j < sp[1]; j++ {
			if !platform[c.rows[j].stop] {
				noStop++
				continue trips
			}
		}
		dir := c.trips.get(tr, "direction_id")
		if dir == "" {
			dir = "0"
		}
		sb.Reset()
		for j := sp[0]; j < sp[1]; j++ {
			if j > sp[0] {
				sb.WriteByte('_')
			}
			sb.WriteString(c.stopIx.strs[c.rows[j].stop])
		}
		sb.WriteString("|" + c.trips.get(tr, "shape_id"))
		sb.WriteString("|" + c.trips.get(tr, "trip_headsign"))
		for j := sp[0]; j < sp[1]; j++ {
			pk, dr := "", ""
			if c.rows[j].pickup != 0 {
				pk = strconv.Itoa(int(c.rows[j].pickup))
			}
			if c.rows[j].drop != 0 {
				dr = strconv.Itoa(int(c.rows[j].drop))
			}
			sb.WriteString("," + pk + "/" + dr + "/" + c.hsIx.strs[c.rows[j].headsign])
		}
		key := c.trips.get(tr, "route_id") + "_" + dir + "_" + md5hex(sb.String())[:8]
		c.tinfos = append(c.tinfos, tinfo{row: tr, tid: tid, key: key, dir: dir})
		if _, ok := patIdx[key]; !ok {
			patIdx[key] = len(c.patterns)
			c.patterns = append(c.patterns, pattern{key: key, first: len(c.tinfos) - 1, trip: int(tid)})
		}
	}
	c.note("trip_without_stop_times", "Trips with no stop_times or no service_id are left out.", skipped)
	c.note("trip_unknown_route_left_out", "Trips whose route_id is not in routes.txt are left out.", noRoute)
	c.note("trip_unknown_stop_left_out", "Trips that call at a stop_id that is not a platform in stops.txt are left out.", noStop)
	c.note("trip_without_end_times_left_out", "Trips whose first or last stop has no time are left out.", noTimes)
}

// fillTimes copies a lone arrival or departure to the other side, and gives untimed stops times interpolated by distance between the timed stops around them, as GTFS asks consumers to do.
func (c *conv) fillTimes() {
	filled := 0
	for _, sp := range c.span {
		if sp[0] < 0 {
			continue
		}
		rows := c.rows[sp[0]:sp[1]]
		for i := range rows {
			if rows[i].arr < 0 {
				rows[i].arr = rows[i].dep
			} else if rows[i].dep < 0 {
				rows[i].dep = rows[i].arr
			}
		}
		prev := -1
		for i := range rows {
			if rows[i].dep < 0 {
				continue
			}
			if prev >= 0 && i-prev > 1 {
				filled += c.interpolate(rows[prev : i+1])
			}
			prev = i
		}
	}
	c.note("times_interpolated", "Untimed stops get times interpolated by distance between the timed stops around them.", filled)
}

// interpolate fills every row between the first and last of rows, by distance, or evenly when the stops have no distance between them.
func (c *conv) interpolate(rows []stRow) int {
	n := len(rows)
	cum := make([]float64, n)
	for i := 1; i < n; i++ {
		cum[i] = cum[i-1] + c.stopDist(rows[i-1].stop, rows[i].stop)
	}
	t0, t1 := rows[0].dep, rows[n-1].arr
	for i := 1; i < n-1; i++ {
		f := float64(i) / float64(n-1)
		if cum[n-1] > 0 {
			f = cum[i] / cum[n-1]
		}
		v := t0 + int32(math.Round(f*float64(t1-t0)))
		rows[i].arr, rows[i].dep = v, v
	}
	return n - 2
}

// stopDist returns the flat-earth distance in degrees of latitude between two stops, or 0 when one is unknown.
func (c *conv) stopDist(a, b uint32) float64 {
	ra, ok1 := c.stopByID[c.stopIx.strs[a]]
	rb, ok2 := c.stopByID[c.stopIx.strs[b]]
	if !ok1 || !ok2 {
		return 0
	}
	lo1, la1 := c.coord(ra)
	lo2, la2 := c.coord(rb)
	return math.Hypot((lo2-lo1)*math.Cos((la1+la2)*math.Pi/360), la2-la1)
}

func (c *conv) loadShapes() error {
	need := map[string]bool{}
	for _, p := range c.patterns {
		if s := c.trips.get(c.tinfos[p.first].row, "shape_id"); s != "" {
			need[s] = true
		}
	}
	c.shapePts = map[string][]float64{}
	f, err := openCSV(c.fs, "shapes.txt")
	if err != nil || f == nil {
		return err
	}
	defer f.close()
	type pt struct {
		seq      int
		lat, lon float64
	}
	tmp := map[string][]pt{}
	n := 0
	for {
		rec, err := f.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		n++
		id := f.get(rec, "shape_id")
		if !need[id] {
			continue
		}
		la, _ := strconv.ParseFloat(strings.TrimSpace(f.get(rec, "shape_pt_lat")), 64)
		lo, _ := strconv.ParseFloat(strings.TrimSpace(f.get(rec, "shape_pt_lon")), 64)
		if _, ok := tmp[id]; !ok {
			id = strings.Clone(id)
		}
		tmp[id] = append(tmp[id], pt{atoiDef(f.get(rec, "shape_pt_sequence"), 0), la, lo})
	}
	c.recs["shapes.txt"] = n
	for id, pts := range tmp {
		sort.SliceStable(pts, func(a, b int) bool { return pts[a].seq < pts[b].seq })
		fl := make([]float64, 0, 2*len(pts))
		for _, p := range pts {
			fl = append(fl, p.lat, p.lon)
		}
		c.shapePts[id] = fl
	}
	return nil
}

func (c *conv) coord(r []string) (float64, float64) {
	la, _ := strconv.ParseFloat(strings.TrimSpace(c.stops.get(r, "stop_lat")), 64)
	lo, _ := strconv.ParseFloat(strings.TrimSpace(c.stops.get(r, "stop_lon")), 64)
	return lo, la
}

// buildLinks makes one ServiceLink per stop pair and shape, holding the piece of the shape between the two stops.
func (c *conv) buildLinks() {
	idx := map[string]int{}
	uncut := 0
	for pi := range c.patterns {
		p := &c.patterns[pi]
		t := &c.tinfos[p.first]
		sp := c.span[p.trip]
		var pieces [][]float64
		shape := c.trips.get(t.row, "shape_id")
		if pts := c.shapePts[shape]; len(pts) >= 4 && sp[1]-sp[0] >= 2 {
			stops := make([][2]float64, 0, sp[1]-sp[0])
			for j := sp[0]; j < sp[1]; j++ {
				r, ok := c.stopByID[c.stopIx.strs[c.rows[j].stop]]
				if !ok {
					break
				}
				lo, la := c.coord(r)
				stops = append(stops, [2]float64{la, lo})
			}
			if len(stops) == sp[1]-sp[0] {
				pieces = cutShape(pts, stops)
			}
			if pieces == nil {
				uncut++
			}
		}
		if pieces == nil {
			shape = ""
		}
		t.shape = shape
		for j := sp[0]; j+1 < sp[1]; j++ {
			f, to := c.ssp(c.stopIx.strs[c.rows[j].stop]), c.ssp(c.stopIx.strs[c.rows[j+1].stop])
			key := f + ">" + to + "|" + shape
			i, ok := idx[key]
			if !ok {
				i = len(c.links)
				idx[key] = i
				l := link{id: c.id("ServiceLink", "SL_"+md5hex(key)), from: f, to: to}
				if pieces != nil {
					l.geom = pieces[j-sp[0]]
				}
				c.links = append(c.links, l)
			}
			p.links = append(p.links, c.links[i].id)
		}
	}
	c.note("shape_not_cut", "Shapes whose stops could not be placed on them in order; their ServiceLinks carry no geometry.", uncut)
}

func (c *conv) loadCalendar() error {
	var crows []calRow
	for _, r := range c.calendar.rows {
		a, ok1 := dayNum(c.calendar.get(r, "start_date"))
		b, ok2 := dayNum(c.calendar.get(r, "end_date"))
		var wd [7]bool
		for i, n := range [7]string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"} {
			wd[i] = strings.TrimSpace(c.calendar.get(r, n)) == "1"
		}
		crows = append(crows, calRow{c.calendar.get(r, "service_id"), wd, a, b, ok1 && ok2})
	}
	drows := map[string][]dateRow{}
	f, err := openCSV(c.fs, "calendar_dates.txt")
	if err != nil {
		return err
	}
	if f != nil {
		defer f.close()
		n := 0
		for {
			rec, err := f.next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return err
			}
			n++
			d, ok := dayNum(f.get(rec, "date"))
			t := strings.TrimSpace(f.get(rec, "exception_type"))
			if !ok || (t != "1" && t != "2") {
				continue
			}
			sid := f.get(rec, "service_id")
			if _, ok := drows[sid]; !ok {
				sid = strings.Clone(sid)
			}
			drows[sid] = append(drows[sid], dateRow{int32(d), t[0]}) //nolint:gosec // day numbers fit
		}
		c.recs["calendar_dates.txt"] = n
	}
	svcIDs := make([]string, 0, len(c.tinfos))
	for _, t := range c.tinfos {
		svcIDs = append(svcIDs, c.trips.get(t.row, "service_id"))
	}
	var rng [2]int
	c.cal, rng = buildCalendar(crows, drows, svcIDs)
	c.calLo, c.calHi = rng[0], rng[1]
	return nil
}

// countLeftOut counts GTFS content that EPIP or this converter does not carry.
func (c *conv) countLeftOut() error {
	files := []struct{ name, msg string }{
		{"pathways.txt", "Pathways are left out: EPIP has no SitePathLink."},
		{"levels.txt", "Levels are not converted yet."},
		{"frequencies.txt", "Frequencies are not converted yet: such trips run only at their stop_times."},
		{"translations.txt", "Translations are not converted yet."},
		{"fare_attributes.txt", "Fares are left out: they are outside EPIP."},
		{"fare_rules.txt", "Fares are left out: they are outside EPIP."},
	}
	for _, fl := range files {
		f, err := openCSV(c.fs, fl.name)
		if err != nil {
			return err
		}
		if f == nil {
			continue
		}
		n := 0
		for {
			_, err := f.next()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				f.close()
				return err
			}
			n++
		}
		f.close()
		c.recs[fl.name] = n
		c.note("not_converted_"+strings.TrimSuffix(fl.name, ".txt"), fl.msg, n)
	}
	blocks, chairNo := 0, 0
	for _, t := range c.tinfos {
		if c.trips.get(t.row, "block_id") != "" {
			blocks++
		}
		if strings.TrimSpace(c.trips.get(t.row, "wheelchair_accessible")) == "2" {
			chairNo++
		}
	}
	c.note("block_left_out", "block_id is left out: EPIP has no BlockRef on ServiceJourney.", blocks)
	c.note("wheelchair_no_left_out", "wheelchair_accessible=2 is left out: EPIP's MobilityFacilityList has no 'not suitable' value.", chairNo)
	other := 0
	for _, r := range c.stops.rows {
		if lt := atoiDef(c.stops.get(r, "location_type"), 0); lt >= 2 {
			other++
		}
	}
	c.note("stop_node_left_out", "Entrances, generic nodes and boarding areas (location_type 2-4) are not converted yet.", other)
	return nil
}

func (c *conv) header() {
	pub := c.cfg.PublicationTime.Format("2006-01-02T15:04:05")
	var fi []string
	if len(c.feedInfo.rows) > 0 {
		fi = c.feedInfo.rows[0]
	}
	fromDate, toDate := "", ""
	fa, okA := dayNum(c.feedInfo.get(fi, "feed_start_date"))
	fb, okB := dayNum(c.feedInfo.get(fi, "feed_end_date"))
	switch {
	case okA && okB && fa <= fb:
		fromDate, toDate = dayStr(fa), dayStr(fb)
	case c.calHi >= 0:
		fromDate, toDate = dayStr(c.calLo), dayStr(c.calHi)
	default:
		fromDate, toDate = c.cfg.PublicationTime.Format("2006-01-02"), c.cfg.PublicationTime.AddDate(1, 0, 0).Format("2006-01-02")
	}
	o := c.o
	o.s(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	o.s(`<PublicationDelivery xmlns="http://www.netex.org.uk/netex" xmlns:gml="http://www.opengis.net/gml/3.2" version="1.1">`)
	o.el("PublicationTimestamp", pub)
	o.el("ParticipantRef", c.cs)
	o.s("<PublicationRequest>")
	o.el("RequestTimestamp", pub)
	o.s("</PublicationRequest><dataObjects>")

	arows := c.agency.rows
	if len(arows) == 0 {
		arows = [][]string{nil}
	}
	c.orgOf = map[string]org{}
	for i, r := range arows {
		k := agencyKey(c.agency.get(r, "agency_id"), c.agency.get(r, "agency_name"), i)
		if r == nil {
			k = "UNKNOWN"
		}
		g := org{k, c.id("Authority", "authority_"+k), c.id("Operator", "operator_"+k)}
		c.orgs = append(c.orgs, g)
		c.orgOf[c.agency.get(r, "agency_id")] = g
	}
	pk := c.orgs[0].key
	o.open("CompositeFrame", c.id("CompositeFrame", pk))
	o.s("<ValidBetween><FromDate>" + fromDate + "T00:00:00</FromDate><ToDate>" + toDate + "T00:00:00</ToDate></ValidBetween>")
	o.frameType("EU_PI_NETWORK_OFFER")
	o.s(`<codespaces><Codespace id="`)
	o.e(c.cs)
	o.s(`"><Xmlns>`)
	o.e(c.cs)
	o.s(`</Xmlns><XmlnsUrl>http://www.`)
	o.e(strings.ToLower(c.cs))
	o.s(`.netex.org</XmlnsUrl><Description>`)
	o.e(c.cs)
	o.s(` NeTEx codespace</Description></Codespace></codespaces><FrameDefaults><DefaultCodespaceRef ref="`)
	o.e(c.cs)
	o.s(`"></DefaultCodespaceRef>`)
	o.s(`<DefaultDataSourceRef ref="`)
	o.e(c.id("DataSource", pk))
	o.s(`"></DefaultDataSourceRef>`)
	o.ref("DefaultResponsibilitySetRef", c.id("ResponsibilitySet", pk))
	o.s("<DefaultLocale>")
	o.el("TimeZone", c.cfg.TimeZone)
	o.el("DefaultLanguage", c.cfg.Language)
	o.s("</DefaultLocale><DefaultLocationSystem>EPSG:4326</DefaultLocationSystem></FrameDefaults><frames>")
}

func (c *conv) resourceFrame() {
	o := c.o
	pk := c.orgs[0].key
	arows := c.agency.rows
	if len(arows) == 0 {
		arows = [][]string{nil}
	}
	o.open("ResourceFrame", c.id("ResourceFrame", pk))
	o.frameType("EU_PI_COMMON")
	o.s("<dataSources>")
	o.open("DataSource", c.id("DataSource", pk))
	o.el("Name", c.agency.get(arows[0], "agency_name"))
	o.s("</DataSource></dataSources><responsibilitySets>")
	o.open("ResponsibilitySet", c.id("ResponsibilitySet", pk))
	o.s("<roles>")
	o.open("ResponsibilityRoleAssignment", c.id("ResponsibilityRoleAssignment", pk))
	o.s("<DataRoleType>creates</DataRoleType>")
	o.ref("ResponsibleOrganisationRef", c.orgs[0].auth)
	o.s("</ResponsibilityRoleAssignment></roles></ResponsibilitySet></responsibilitySets><organisations>")
	for pass, typ := range []string{"Authority", "Operator"} {
		for i, r := range arows {
			g := c.orgs[i]
			id := g.auth
			if pass == 1 {
				id = g.oper
			}
			name := c.agency.get(r, "agency_name")
			o.open(typ, id)
			o.el("PublicCode", g.key)
			o.el("Name", name)
			o.el("ShortName", name)
			o.el("LegalName", name)
			o.s("<ContactDetails>")
			o.elOpt("Email", c.agency.get(r, "agency_email"))
			o.elOpt("Phone", c.agency.get(r, "agency_phone"))
			o.elOpt("Url", c.agency.get(r, "agency_url"))
			o.s("</ContactDetails>")
			o.el("OrganisationType", strings.ToLower(typ))
			o.s("</" + typ + ">")
			c.made[typ]++
		}
	}
	o.s("</organisations></ResourceFrame>")
}

// calendarFrame writes each service as a DayType with its weekdays, a UicOperatingPeriod with the exact days, and dated exceptions.
func (c *conv) calendarFrame() {
	o := c.o
	pk := c.orgs[0].key
	from, to := "", ""
	if c.calHi >= 0 {
		from, to = dayStr(c.calLo), dayStr(c.calHi)
	} else {
		from, to = c.cfg.PublicationTime.Format("2006-01-02"), c.cfg.PublicationTime.AddDate(1, 0, 0).Format("2006-01-02")
	}
	o.open("ServiceCalendarFrame", c.id("ServiceCalendarFrame", pk))
	o.frameType("EU_PI_CALENDAR")
	o.open("ServiceCalendar", c.id("ServiceCalendar", "ALL_SERVICES"))
	o.s("<FromDate>" + from + "</FromDate><ToDate>" + to + "</ToDate><dayTypes>")
	for _, v := range c.cal {
		o.open("DayType", c.id("DayType", "day_type_"+v.id))
		if v.hasRange {
			o.s("<properties><PropertyOfDay><DaysOfWeek>" + daysOfWeek(v.weekdays) + "</DaysOfWeek></PropertyOfDay></properties>")
		}
		o.s("</DayType>")
		c.made["DayType"]++
	}
	o.s("</dayTypes>")
	started := false
	for _, v := range c.cal {
		if !v.hasRange {
			continue
		}
		if !started {
			o.s("<operatingPeriods>")
			started = true
		}
		o.open("UicOperatingPeriod", c.id("OperatingPeriod", "operating_period_"+v.id))
		o.s("<FromDate>" + dayStr(v.lo) + "T00:00:00</FromDate><ToDate>" + dayStr(v.hi) + "T00:00:00</ToDate><ValidDayBits>")
		o.b(v.bits)
		o.s("</ValidDayBits></UicOperatingPeriod>")
		c.made["OperatingPeriod"]++
	}
	if started {
		o.s("</operatingPeriods>")
	}
	o.s("<dayTypeAssignments>")
	for _, v := range c.cal {
		n := 0
		head := func() {
			n++
			o.s(`<DayTypeAssignment id="`)
			o.e(c.id("DayTypeAssignment", "day_type_assignment_"+v.id+"_"+strconv.Itoa(n)))
			o.s(`" order="` + strconv.Itoa(n) + `" version="1">`)
		}
		if v.hasRange {
			head()
			o.ref("OperatingPeriodRef", c.id("OperatingPeriod", "operating_period_"+v.id))
			o.ref("DayTypeRef", c.id("DayType", "day_type_"+v.id))
			o.s("<isAvailable>true</isAvailable></DayTypeAssignment>")
		}
		for _, e := range v.except {
			head()
			o.s("<Date>" + dayStr(e.day) + "</Date>")
			o.ref("DayTypeRef", c.id("DayType", "day_type_"+v.id))
			o.s("<isAvailable>" + strconv.FormatBool(e.available) + "</isAvailable></DayTypeAssignment>")
		}
	}
	o.s("</dayTypeAssignments></ServiceCalendar></ServiceCalendarFrame>")
}

func (c *conv) routeType(rid string) int {
	if r, ok := c.routeByID[rid]; ok {
		return atoiDef(c.routes.get(r, "route_type"), 0)
	}
	return 3
}

func (c *conv) ddID(text string) string { return c.id("DestinationDisplay", "DD_"+md5hex(text)) }

type assign struct {
	stop  []string
	place string
}

// assignments lists every platform with the StopPlace that holds it: a station for its children, itself for a lone stop.
func (c *conv) assignments() []assign {
	var as []assign
	for _, r := range c.stops.rows {
		lt := atoiDef(c.stops.get(r, "location_type"), 0)
		sid := c.stops.get(r, "stop_id")
		if lt == 1 {
			for _, k := range c.children[sid] {
				as = append(as, assign{k, c.id("StopPlace", "stop_place_"+sid)})
			}
		} else if c.lone(r) {
			as = append(as, assign{r, c.id("StopPlace", "stop_place_"+sid)})
		}
	}
	return as
}

func (c *conv) serviceFrame() {
	o := c.o
	pk := c.orgs[0].key
	o.open("ServiceFrame", c.id("ServiceFrame", pk))
	o.frameType("EU_PI_NETWORK")
	o.s("<directions>")
	for _, d := range []string{"0", "1"} {
		o.open("Direction", c.id("Direction", "direction_"+d))
		o.s("</Direction>")
	}
	o.s("</directions>")
	// One Route per route and direction, from its longest pattern.
	best := map[string]int{}
	var rdOrder []string
	for pi, p := range c.patterns {
		t := c.tinfos[p.first]
		rd := c.trips.get(t.row, "route_id") + "\x00" + t.dir
		if b, ok := best[rd]; !ok {
			best[rd] = pi
			rdOrder = append(rdOrder, rd)
		} else {
			sb, sp := c.span[c.patterns[b].trip], c.span[p.trip]
			if sp[1]-sp[0] > sb[1]-sb[0] {
				best[rd] = pi
			}
		}
	}
	sort.Strings(rdOrder)
	o.s("<routePoints>")
	rpSeen := map[uint32]bool{}
	for _, rd := range rdOrder {
		sp := c.span[c.patterns[best[rd]].trip]
		for j := sp[0]; j < sp[1]; j++ {
			if rpSeen[c.rows[j].stop] {
				continue
			}
			rpSeen[c.rows[j].stop] = true
			sid := c.stopIx.strs[c.rows[j].stop]
			o.open("RoutePoint", c.id("RoutePoint", "RP_"+sid))
			if r, ok := c.stopByID[sid]; ok {
				o.loc(c.coord(r))
			} else {
				o.loc(0, 0)
			}
			o.s("</RoutePoint>")
		}
	}
	o.s("</routePoints><routes>")
	routesOfLine := map[string][]string{}
	for _, rd := range rdOrder {
		parts := strings.SplitN(rd, "\x00", 2)
		rid, dir := parts[0], parts[1]
		id := c.id("Route", "route_"+rid+"_"+dir)
		routesOfLine[rid] = append(routesOfLine[rid], id)
		o.open("Route", id)
		o.ref("LineRef", c.id("Line", "line_"+rid))
		o.ref("DirectionRef", c.id("Direction", "direction_"+dir))
		o.s("<pointsInSequence>")
		sp := c.span[c.patterns[best[rd]].trip]
		for j := sp[0]; j < sp[1]; j++ {
			n := strconv.Itoa(j - sp[0] + 1)
			o.s(`<PointOnRoute id="`)
			o.e(c.id("PointOnRoute", "POR_"+rid+"_"+dir+"_"+n))
			o.s(`" order="` + n + `" version="1">`)
			o.ref("RoutePointRef", c.id("RoutePoint", "RP_"+c.stopIx.strs[c.rows[j].stop]))
			o.s("</PointOnRoute>")
		}
		o.s("</pointsInSequence></Route>")
		c.made["Route"]++
	}
	o.s("</routes><lines>")
	noTrips, unknownType := 0, 0
	for _, r := range c.routes.rows {
		rid := c.routes.get(r, "route_id")
		refs := routesOfLine[rid]
		if len(refs) == 0 {
			noTrips++
			continue
		}
		if _, ok := modeOK(atoiDef(c.routes.get(r, "route_type"), 3)); !ok {
			unknownType++
		}
		g, ok := c.orgOf[c.routes.get(r, "agency_id")]
		if !ok {
			g = c.orgs[0]
		}
		o.open("Line", c.id("Line", "line_"+rid))
		o.s("<keyList><KeyValue><Key>gtfs_route_id</Key>")
		o.el("Value", rid)
		o.s("</KeyValue></keyList>")
		o.el("Name", c.routes.get(r, "route_long_name"))
		o.elOpt("ShortName", c.routes.get(r, "route_short_name"))
		o.elOpt("Description", c.routes.get(r, "route_desc"))
		o.el("TransportMode", mode(atoiDef(c.routes.get(r, "route_type"), 3)))
		o.elOpt("Url", c.routes.get(r, "route_url"))
		o.el("PublicCode", c.routes.get(r, "route_short_name"))
		o.el("PrivateCode", rid)
		o.ref("AuthorityRef", g.auth)
		o.s("<additionalOperators>")
		o.ref("OperatorRef", g.oper)
		o.s("</additionalOperators><routes>")
		for _, x := range refs {
			o.ref("RouteRef", x)
		}
		o.s("</routes>")
		if c.routes.get(r, "route_color") != "" || c.routes.get(r, "route_text_color") != "" {
			o.s("<Presentation>")
			o.elOpt("Colour", c.routes.get(r, "route_color"))
			o.elOpt("TextColour", c.routes.get(r, "route_text_color"))
			o.s("</Presentation>")
		}
		o.s("<allowedDirections>")
		for _, d := range []string{"0", "1"} {
			o.s(`<AllowedLineDirection id="`)
			o.e("allowed_direction_" + rid + "_" + d)
			o.s(`" version="1">`)
			o.ref("DirectionRef", c.id("Direction", "direction_"+d))
			o.s("</AllowedLineDirection>")
		}
		o.s("</allowedDirections></Line>")
		c.made["Line"]++
	}
	o.s("</lines>")
	c.note("line_without_trips_left_out", "Routes with no trips are left out, as nothing would reference their Line.", noTrips)
	c.note("route_type_unknown", "Routes whose route_type has no EPIP mode, such as taxi or an unknown code, are written as bus.", unknownType)
	c.destinationDisplays()
	as := c.assignments()
	o.s("<scheduledStopPoints>")
	for _, a := range as {
		sid := c.stops.get(a.stop, "stop_id")
		o.open("ScheduledStopPoint", c.ssp(sid))
		o.el("Name", c.stops.get(a.stop, "stop_name"))
		o.loc(c.coord(a.stop))
		o.elOpt("PublicCode", c.stops.get(a.stop, "stop_code"))
		o.s("</ScheduledStopPoint>")
		c.made["ScheduledStopPoint"]++
	}
	o.s("</scheduledStopPoints>")
	c.serviceLinks()
	c.connections()
	o.s("<stopAssignments>")
	for _, a := range as {
		sid := c.stops.get(a.stop, "stop_id")
		o.s(`<PassengerStopAssignment id="`)
		o.e(c.id("PassengerStopAssignment", "assignment_"+sid))
		o.s(`" order="1" version="1">`)
		o.ref("ScheduledStopPointRef", c.ssp(sid))
		o.ref("StopPlaceRef", a.place)
		o.ref("QuayRef", c.id("Quay", "quay_"+sid))
		o.s("</PassengerStopAssignment>")
	}
	o.s("</stopAssignments>")
	c.journeyPatterns()
	o.s("</ServiceFrame>")
}

func (c *conv) destinationDisplays() {
	var dds []string
	seen := map[string]bool{}
	add := func(t string) {
		if t != "" && !seen[t] {
			seen[t] = true
			dds = append(dds, t)
		}
	}
	for _, p := range c.patterns {
		sp := c.span[p.trip]
		for j := sp[0]; j < sp[1]; j++ {
			add(c.hsIx.strs[c.rows[j].headsign])
		}
		add(c.trips.get(c.tinfos[p.first].row, "trip_headsign"))
	}
	if len(dds) == 0 {
		return
	}
	c.o.s("<destinationDisplays>")
	for _, t := range dds {
		c.o.open("DestinationDisplay", c.ddID(t))
		c.o.el("FrontText", t)
		c.o.s("</DestinationDisplay>")
	}
	c.o.s("</destinationDisplays>")
}

func (c *conv) serviceLinks() {
	if len(c.links) == 0 {
		return
	}
	o := c.o
	o.s("<serviceLinks>")
	buf := make([]byte, 0, 1<<12)
	for _, l := range c.links {
		o.open("ServiceLink", l.id)
		if len(l.geom) >= 4 {
			o.s(`<gml:LineString gml:id="` + gmlID(l.id) + `" srsName="EPSG:4326"><gml:posList>`)
			buf = appendPosList(buf[:0], l.geom)
			o.b(buf)
			o.s("</gml:posList></gml:LineString>")
			c.made["ServiceLink with geometry"]++
		}
		o.ref("FromPointRef", l.from)
		o.ref("ToPointRef", l.to)
		o.s("</ServiceLink>")
		c.made["ServiceLink"]++
	}
	o.s("</serviceLinks>")
}

// platforms returns the platforms a transfer's stop_id stands for: itself, or every platform of a station.
func (c *conv) platforms(id string) []string {
	switch c.locType(id) {
	case 0:
		return []string{id}
	case 1:
		ps := make([]string, 0, len(c.children[id]))
		for _, r := range c.children[id] {
			ps = append(ps, c.stops.get(r, "stop_id"))
		}
		return ps
	}
	return nil
}

// connections writes stop-to-stop transfers, platform rows before station rows so the more exact one wins; trip-to-trip ones go to the TimetableFrame.
func (c *conv) connections() {
	o := c.o
	type conn struct {
		f, t string
		secs int
	}
	var exact, fromStation []conn
	notPossible, routeLevel, unknown := 0, 0, 0
	for _, r := range c.transfers.rows {
		if c.transfers.get(r, "from_trip_id") != "" && c.transfers.get(r, "to_trip_id") != "" {
			continue
		}
		if c.transfers.get(r, "from_route_id") != "" || c.transfers.get(r, "to_route_id") != "" {
			routeLevel++
			continue
		}
		typ := atoiDef(c.transfers.get(r, "transfer_type"), 0)
		if typ == 3 {
			notPossible++
			continue
		}
		from, to := c.transfers.get(r, "from_stop_id"), c.transfers.get(r, "to_stop_id")
		fs, ts := c.platforms(from), c.platforms(to)
		if len(fs) == 0 || len(ts) == 0 {
			unknown++
			continue
		}
		secs := atoiDef(c.transfers.get(r, "min_transfer_time"), 0)
		for _, f := range fs {
			for _, t := range ts {
				if c.locType(from) == 0 && c.locType(to) == 0 {
					exact = append(exact, conn{f, t, secs})
				} else {
					fromStation = append(fromStation, conn{f, t, secs})
				}
			}
		}
	}
	started := false
	seen := map[string]bool{}
	for _, cn := range append(exact, fromStation...) {
		f, t := c.ssp(cn.f), c.ssp(cn.t)
		key := md5hex(f + ">" + t)
		if seen[key] {
			continue
		}
		seen[key] = true
		if !started {
			o.s("<connections>")
			started = true
		}
		secs := cn.secs
		o.open("Connection", c.id("Connection", "CONN_"+key))
		o.s("<WalkTransferDuration><DefaultDuration>PT" + strconv.Itoa(secs) + "S</DefaultDuration></WalkTransferDuration><From>")
		o.ref("ScheduledStopPointRef", f)
		o.s("</From><To>")
		o.ref("ScheduledStopPointRef", t)
		o.s("</To></Connection>")
		c.made["Connection"]++
	}
	if started {
		o.s("</connections>")
	}
	c.note("transfer_not_possible_left_out", "transfer_type=3 is left out: EPIP has no TransferRestriction.", notPossible)
	c.note("transfer_unknown_stop_left_out", "Transfers whose stop is not a platform or station in stops.txt are left out.", unknown)
	c.note("route_transfer_left_out", "Route-to-route transfers are left out: EPIP has no InterchangeRule.", routeLevel)
}

func (c *conv) journeyPatterns() {
	o := c.o
	o.s("<journeyPatterns>")
	for _, p := range c.patterns {
		t := c.tinfos[p.first]
		rid := c.trips.get(t.row, "route_id")
		o.open("ServiceJourneyPattern", c.id("JourneyPattern", "JP_"+p.key))
		// EPIP allows RouteRef or RouteView here; RouteView puts the line on the pattern, as NAPCORE asks.
		o.s("<RouteView>")
		o.ref("LineRef", c.id("Line", "line_"+rid))
		o.s("</RouteView>")
		o.ref("DirectionRef", c.id("Direction", "direction_"+t.dir))
		if hs := c.trips.get(t.row, "trip_headsign"); hs != "" {
			o.ref("DestinationDisplayRef", c.ddID(hs))
		}
		o.s("<pointsInSequence>")
		sp := c.span[p.trip]
		for j := sp[0]; j < sp[1]; j++ {
			i := j - sp[0]
			o.s(`<StopPointInJourneyPattern id="`)
			o.e(c.id("StopPointInJourneyPattern", "JP_"+p.key+"_"+strconv.Itoa(i)))
			o.s(`" order="` + strconv.Itoa(i+1) + `" version="1">`)
			o.ref("ScheduledStopPointRef", c.ssp(c.stopIx.strs[c.rows[j].stop]))
			if i < len(p.links) {
				o.ref("OnwardServiceLinkRef", p.links[i])
			}
			row := c.rows[j]
			o.s("<ForAlighting>" + strconv.FormatBool(row.drop != 1) + "</ForAlighting><ForBoarding>" + strconv.FormatBool(row.pickup != 1) + "</ForBoarding>")
			if row.headsign != 0 {
				o.ref("DestinationDisplayRef", c.ddID(c.hsIx.strs[row.headsign]))
			}
			if row.pickup == 2 || row.pickup == 3 || row.drop == 2 || row.drop == 3 {
				o.s("<RequestStop>true</RequestStop>")
			}
			o.s("</StopPointInJourneyPattern>")
		}
		o.s("</pointsInSequence></ServiceJourneyPattern>")
		c.made["ServiceJourneyPattern"]++
	}
	o.s("</journeyPatterns>")
}

func (c *conv) siteFrame() {
	o := c.o
	pk := c.orgs[0].key
	stopMode := map[string]string{}
	for _, t := range c.tinfos {
		m := mode(c.routeType(c.trips.get(t.row, "route_id")))
		sp := c.span[t.tid]
		for j := sp[0]; j < sp[1]; j++ {
			sid := c.stopIx.strs[c.rows[j].stop]
			if _, ok := stopMode[sid]; !ok {
				stopMode[sid] = m
			}
		}
	}
	children := c.children
	o.open("SiteFrame", c.id("SiteFrame", pk))
	o.frameType("EU_PI_STOP")
	o.s("<stopPlaces>")
	quay := func(r []string) {
		sid := c.stops.get(r, "stop_id")
		o.open("Quay", c.id("Quay", "quay_"+sid))
		o.el("Name", c.stops.get(r, "stop_name"))
		o.s("<Centroid>")
		o.loc(c.coord(r))
		o.s("</Centroid>")
		if v := map[string]string{"1": "true", "2": "false"}[strings.TrimSpace(c.stops.get(r, "wheelchair_boarding"))]; v != "" {
			o.open("AccessibilityAssessment", c.id("AccessibilityAssessment", "Q_"+sid))
			o.s("<MobilityImpairedAccess>" + v + "</MobilityImpairedAccess><limitations><AccessibilityLimitation><WheelchairAccess>" + v +
				"</WheelchairAccess></AccessibilityLimitation></limitations></AccessibilityAssessment>")
		}
		o.elOpt("PublicCode", c.stops.get(r, "platform_code"))
		o.s("</Quay>")
		c.made["Quay"]++
	}
	unserved := 0
	place := func(r []string, typ, m string, quays [][]string) {
		sid := c.stops.get(r, "stop_id")
		o.open("StopPlace", c.id("StopPlace", "stop_place_"+sid))
		o.el("Name", c.stops.get(r, "stop_name"))
		o.s("<Centroid>")
		o.loc(c.coord(r))
		o.s("</Centroid>")
		o.elOpt("Url", c.stops.get(r, "stop_url"))
		o.elOpt("TransportMode", m)
		if typ == "" {
			typ = "other"
			unserved++
		}
		o.el("StopPlaceType", typ)
		o.s("<quays>")
		for _, q := range quays {
			quay(q)
		}
		o.s("</quays></StopPlace>")
		c.made["StopPlace"]++
	}
	for _, r := range c.stops.rows {
		lt := atoiDef(c.stops.get(r, "location_type"), 0)
		sid := c.stops.get(r, "stop_id")
		if lt == 1 && len(children[sid]) > 0 {
			m := ""
			for _, k := range children[sid] {
				if m = stopMode[c.stops.get(k, "stop_id")]; m != "" {
					break
				}
			}
			place(r, stationTypeOfMode[m], m, children[sid])
		} else if c.lone(r) {
			m := stopMode[sid]
			place(r, stopPlaceTypeOfMode[m], m, [][]string{r})
		}
	}
	o.s("</stopPlaces></SiteFrame>")
	c.note("stop_without_trips", "Stops no trip serves get StopPlaceType 'other' and no TransportMode.", unserved)
}

// timetableFrame streams one ServiceJourney per trip.
func (c *conv) timetableFrame() error {
	o := c.o
	pk := c.orgs[0].key
	o.open("TimetableFrame", c.id("TimetableFrame", pk))
	o.frameType("EU_PI_TIMETABLE")
	o.s("<vehicleJourneys>")
	jb := make([]byte, 0, 1<<14)
	for n, t := range c.tinfos {
		if n%1024 == 0 {
			if err := c.alive(); err != nil {
				return err
			}
		}
		tripID := c.trips.get(t.row, "trip_id")
		rid := c.trips.get(t.row, "route_id")
		rt := c.routeType(rid)
		sp := c.span[t.tid]
		first, last := c.rows[sp[0]], c.rows[sp[1]-1]
		dep, depOff := hms(first.dep)
		jb = jb[:0]
		jb = append(jb, `<ServiceJourney id="`...)
		jb = escape(jb, c.id("ServiceJourney", "TRIP_"+tripID))
		jb = append(jb, `" version="1">`...)
		if name := c.trips.get(t.row, "trip_short_name"); name != "" {
			jb = append(jb, "<Name>"...)
			jb = escape(jb, name)
			jb = append(jb, "</Name>"...)
		}
		jb = append(jb, "<TransportMode>"...)
		jb = append(jb, mode(rt)...)
		jb = append(jb, "</TransportMode><DepartureTime>"...)
		jb = append(jb, dep...)
		jb = append(jb, "</DepartureTime>"...)
		if depOff > 0 {
			jb = append(jb, "<DepartureDayOffset>"...)
			jb = strconv.AppendInt(jb, int64(depOff), 10)
			jb = append(jb, "</DepartureDayOffset>"...)
		}
		durS := int(last.arr) - int(first.dep)
		if last.arr < 0 || first.dep < 0 || sp[1]-sp[0] < 2 {
			durS = 0
		}
		if durS < 0 {
			durS += 86400
		}
		jb = append(jb, fmt.Sprintf("<JourneyDuration>PT%dH%dM", durS/3600, durS/60%60)...)
		if durS%60 != 0 {
			jb = append(jb, fmt.Sprintf("%dS", durS%60)...)
		}
		jb = append(jb, "</JourneyDuration>"...)
		jb = append(jb, `<dayTypes><DayTypeRef ref="`...)
		jb = escape(jb, c.id("DayType", "day_type_"+c.trips.get(t.row, "service_id")))
		jb = append(jb, `" version="1"></DayTypeRef></dayTypes><ServiceJourneyPatternRef ref="`...)
		jb = escape(jb, c.id("JourneyPattern", "JP_"+t.key))
		jb = append(jb, `" version="1"></ServiceJourneyPatternRef><VehicleTypeRef ref="`...)
		jb = escape(jb, c.id("VehicleType", "vehicle_type_"+strconv.Itoa(rt)))
		jb = append(jb, `" version="1"></VehicleTypeRef><passingTimes>`...)
		eTrip := string(escape(nil, tripID))
		eKey := string(escape(nil, t.key))
		eCs := string(escape(nil, c.cs))
		for j := sp[0]; j < sp[1]; j++ {
			i := j - sp[0]
			jb = append(jb, `<TimetabledPassingTime id="`...)
			jb = append(jb, eCs...)
			jb = append(jb, `::TimetabledPassingTime:TPT_`...)
			jb = append(jb, eTrip...)
			jb = append(jb, '_')
			jb = strconv.AppendInt(jb, int64(i), 10)
			jb = append(jb, `::" version="1"><StopPointInJourneyPatternRef ref="`...)
			jb = append(jb, eCs...)
			jb = append(jb, `::StopPointInJourneyPattern:JP_`...)
			jb = append(jb, eKey...)
			jb = append(jb, '_')
			jb = strconv.AppendInt(jb, int64(i), 10)
			jb = append(jb, `::" version="1"></StopPointInJourneyPatternRef>`...)
			for k, v := range []int32{c.rows[j].arr, c.rows[j].dep} {
				if v < 0 {
					continue
				}
				tag, otag := "ArrivalTime", "ArrivalDayOffset"
				if k == 1 {
					tag, otag = "DepartureTime", "DepartureDayOffset"
				}
				s, off := hms(v)
				jb = append(jb, "<"+tag+">"+s+"</"+tag+">"...)
				if off > 0 {
					jb = append(jb, "<"+otag+">"...)
					jb = strconv.AppendInt(jb, int64(off), 10)
					jb = append(jb, "</"+otag+">"...)
				}
			}
			jb = append(jb, "</TimetabledPassingTime>"...)
		}
		jb = append(jb, "</passingTimes>"...)
		chair := strings.TrimSpace(c.trips.get(t.row, "wheelchair_accessible")) == "1"
		bikes := map[string]string{"1": "cyclesAllowed", "2": "noCycles"}[strings.TrimSpace(c.trips.get(t.row, "bikes_allowed"))]
		if chair || bikes != "" {
			jb = append(jb, `<facilities><ServiceFacilitySet id="`...)
			jb = escape(jb, c.id("ServiceFacilitySet", "SJ_"+tripID))
			jb = append(jb, `" version="1">`...)
			if chair {
				jb = append(jb, "<MobilityFacilityList>suitableForWheelchairs</MobilityFacilityList>"...)
			}
			if bikes != "" {
				jb = append(jb, "<LuggageCarriageFacilityList>"+bikes+"</LuggageCarriageFacilityList>"...)
			}
			jb = append(jb, "</ServiceFacilitySet></facilities>"...)
		}
		jb = append(jb, "</ServiceJourney>"...)
		o.b(jb)
		c.made["ServiceJourney"]++
	}
	o.s("</vehicleJourneys>")
	c.interchanges()
	o.s("<vehicleTypes>")
	var rts []int
	seenRT := map[int]bool{}
	for _, r := range c.routes.rows {
		rt := atoiDef(c.routes.get(r, "route_type"), 0)
		if !seenRT[rt] {
			seenRT[rt] = true
			rts = append(rts, rt)
		}
	}
	sort.Ints(rts)
	for _, rt := range rts {
		o.open("VehicleType", c.id("VehicleType", "vehicle_type_"+strconv.Itoa(rt)))
		o.el("Name", vehicleName(rt))
		o.s("</VehicleType>")
	}
	o.s("</vehicleTypes></TimetableFrame>")
	return nil
}

// interchanges writes trip-to-trip transfers; EPIP carries StaySeated and StandardTransferTime but not Guaranteed.
func (c *conv) interchanges() {
	o := c.o
	started := false
	guaranteed, unknownTrip, unknownStop := 0, 0, 0
	written := make([]bool, len(c.tripIx.strs))
	for _, t := range c.tinfos {
		written[t.tid] = true
	}
	seen := map[string]bool{}
	for _, r := range c.transfers.rows {
		ft, tt := c.transfers.get(r, "from_trip_id"), c.transfers.get(r, "to_trip_id")
		typ := atoiDef(c.transfers.get(r, "transfer_type"), 0)
		if ft == "" || tt == "" || typ == 3 {
			continue
		}
		fi, ok1 := c.tripIx.ids[ft]
		ti, ok2 := c.tripIx.ids[tt]
		if !ok1 || !ok2 || !written[fi] || !written[ti] {
			unknownTrip++
			continue
		}
		fp, tp := c.pointOn(c.transfers.get(r, "from_stop_id"), fi, true), c.pointOn(c.transfers.get(r, "to_stop_id"), ti, false)
		if fp == "" || tp == "" {
			unknownStop++
			continue
		}
		id := c.id("ServiceJourneyInterchange", ft+"_"+tt+"_"+fp+"_"+tp)
		if seen[id] {
			continue
		}
		seen[id] = true
		if !started {
			o.s("<journeyInterchanges>")
			started = true
		}
		o.open("ServiceJourneyInterchange", id)
		switch typ {
		case 1:
			guaranteed++
		case 4:
			o.s("<StaySeated>true</StaySeated>")
		case 5:
			o.s("<StaySeated>false</StaySeated>")
		}
		if secs := atoiDef(c.transfers.get(r, "min_transfer_time"), 0); typ == 2 && secs > 0 {
			o.s("<StandardTransferTime>PT" + strconv.Itoa(secs) + "S</StandardTransferTime>")
		}
		o.ref("FromPointRef", c.ssp(fp))
		o.ref("ToPointRef", c.ssp(tp))
		o.s(`<FromJourneyRef ref="`)
		o.e(c.id("ServiceJourney", "TRIP_"+ft))
		o.s(`"></FromJourneyRef><ToJourneyRef ref="`)
		o.e(c.id("ServiceJourney", "TRIP_"+tt))
		o.s(`"></ToJourneyRef></ServiceJourneyInterchange>`)
		c.made["ServiceJourneyInterchange"]++
	}
	if started {
		o.s("</journeyInterchanges>")
	}
	c.note("guaranteed_left_out", "transfer_type=1 between trips keeps its interchange but loses the guarantee: EPIP has no Guaranteed.", guaranteed)
	c.note("interchange_unknown_trip_left_out", "Trip-to-trip transfers whose trip is not in the output are left out.", unknownTrip)
	c.note("interchange_unknown_stop_left_out", "Trip-to-trip transfers whose stop the trip does not reach are left out.", unknownStop)
}

// pointOn returns the platform where a trip meets a transfer's stop_id: the trip's last stop (from side) or first stop (to side) when the stop_id is empty, the stop itself for a platform, the platform the trip uses for a station, and "" otherwise.
func (c *conv) pointOn(stop string, tid uint32, last bool) string {
	sp := c.span[tid]
	if stop == "" {
		j := sp[0]
		if last {
			j = sp[1] - 1
		}
		return c.stopIx.strs[c.rows[j].stop]
	}
	switch c.locType(stop) {
	case 0:
		return stop
	case 1:
		for k := 0; k < sp[1]-sp[0]; k++ {
			j := sp[0] + k
			if last {
				j = sp[1] - 1 - k
			}
			id := c.stopIx.strs[c.rows[j].stop]
			if c.stops.get(c.stopByID[id], "parent_station") == stop {
				return id
			}
		}
	}
	return ""
}

func (c *conv) report() *converter.ConversionReport {
	end := time.Now()
	dur := end.Sub(c.start).Seconds()
	total := 0
	for _, n := range c.recs {
		total += n
	}
	entities := 0
	for k, n := range c.made {
		if k != "ServiceLink with geometry" {
			entities += n
		}
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	rep := &converter.ConversionReport{
		Summary: converter.ConversionSummary{Status: "success", StartTime: c.start, EndTime: end, DurationSeconds: dur,
			ParticipantRef: c.cs, ConverterVersion: Version},
		InputStatistics: converter.InputStatistics{FilesProcessed: len(c.recs), TotalRecords: total, RecordsByFile: c.recs,
			Agencies: c.recs["agency.txt"], Stops: c.recs["stops.txt"], Routes: c.recs["routes.txt"], Trips: c.recs["trips.txt"],
			StopTimes: c.recs["stop_times.txt"], CalendarDates: c.recs["calendar_dates.txt"], Shapes: c.recs["shapes.txt"],
			Transfers: c.recs["transfers.txt"], FareAttributes: c.recs["fare_attributes.txt"], Pathways: c.recs["pathways.txt"]},
		OutputStatistics: converter.OutputStatistics{TotalEntities: entities, EntitiesByType: c.made,
			Authorities: c.made["Authority"], Operators: c.made["Operator"], StopPlaces: c.made["StopPlace"], Quays: c.made["Quay"],
			ScheduledStopPoints: c.made["ScheduledStopPoint"], Lines: c.made["Line"], Routes: c.made["Route"],
			ServiceJourneys: c.made["ServiceJourney"], JourneyPatterns: c.made["ServiceJourneyPattern"], DayTypes: c.made["DayType"],
			OperatingPeriods: c.made["OperatingPeriod"], ServiceLinks: c.made["ServiceLink"]},
		Performance:   converter.PerformanceMetrics{TotalDurationSeconds: dur, PeakMemoryUsageMB: float64(ms.HeapSys) / 1e6},
		EntityMapping: map[string]converter.EntityMappingStats{},
		Warnings:      []converter.Warning{},
		Errors:        []converter.ConversionError{},
	}
	if dur > 0 {
		rep.Performance.RecordsPerSecond = float64(total) / dur
	}
	keys := make([]string, 0, len(c.warn))
	for k := range c.warn {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		rep.Warnings = append(rep.Warnings, *c.warn[k])
	}
	return rep
}
