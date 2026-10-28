package stream

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
)

// csvFile is one open GTFS file read row by row.
type csvFile struct {
	name string
	r    *csv.Reader
	c    io.Closer
	col  map[string]int
}

// openCSV opens a GTFS file by name, also inside one sub-folder; it returns nil when the file is absent.
func openCSV(fsys fs.FS, name string) (*csvFile, error) {
	rc, err := fsys.Open(name)
	if errors.Is(err, fs.ErrNotExist) {
		if m, _ := fs.Glob(fsys, "*/"+name); len(m) > 0 {
			rc, err = fsys.Open(m[0])
		}
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	r := csv.NewReader(bufio.NewReaderSize(rc, 1<<16))
	r.FieldsPerRecord = -1
	r.ReuseRecord = true
	hdr, err := r.Read()
	if errors.Is(err, io.EOF) {
		_ = rc.Close()
		return nil, nil
	}
	if err != nil {
		_ = rc.Close()
		return nil, fmt.Errorf("read %s header: %w", name, err)
	}
	col := map[string]int{}
	for i, h := range hdr {
		col[strings.TrimPrefix(strings.TrimSpace(h), "\ufeff")] = i
	}
	return &csvFile{name: name, r: r, c: rc, col: col}, nil
}

// next returns the next row, valid until the following call, or io.EOF.
func (f *csvFile) next() ([]string, error) {
	rec, err := f.r.Read()
	if err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("read %s: %w", f.name, err)
	}
	return rec, err
}

func (f *csvFile) get(rec []string, c string) string {
	if i, ok := f.col[c]; ok && i < len(rec) {
		return rec[i]
	}
	return ""
}

func (f *csvFile) close() { _ = f.c.Close() }

// table is one small GTFS file held as rows of strings.
type table struct {
	col  map[string]int
	rows [][]string
}

func (t *table) get(r []string, c string) string {
	if i, ok := t.col[c]; ok && i < len(r) {
		return r[i]
	}
	return ""
}

func readTable(fsys fs.FS, name string) (*table, error) {
	f, err := openCSV(fsys, name)
	t := &table{col: map[string]int{}}
	if f == nil || err != nil {
		return t, err
	}
	defer f.close()
	t.col = f.col
	for {
		rec, err := f.next()
		if errors.Is(err, io.EOF) {
			return t, nil
		}
		if err != nil {
			return nil, err
		}
		t.rows = append(t.rows, append([]string(nil), rec...))
	}
}

type interner struct {
	ids  map[string]uint32
	strs []string
}

func newInterner() *interner { return &interner{ids: map[string]uint32{}} }

func (in *interner) id(s string) uint32 {
	if v, ok := in.ids[s]; ok {
		return v
	}
	v := uint32(len(in.strs)) //nolint:gosec // a feed has far fewer than 4 billion distinct IDs
	s = strings.Clone(s)
	in.ids[s] = v
	in.strs = append(in.strs, s)
	return v
}

// stRow is one stop_times row in compact form.
type stRow struct {
	trip, stop, seq, headsign uint32
	arr, dep                  int32
	pickup, drop              uint8
}

// parseTime returns seconds after midnight, or -1 for an empty time.
func parseTime(s string) int32 {
	s = strings.TrimSpace(s)
	if s == "" {
		return -1
	}
	var v, part int32
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' {
			v = v*60 + part
			part = 0
			continue
		}
		part = part*10 + int32(c-'0')
	}
	return v*60 + part
}

func atoiDef(s string, d int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
		return v
	}
	return d
}
