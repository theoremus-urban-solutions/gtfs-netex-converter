package stream

import (
	"sort"
	"strings"
	"time"
)

func dayNum(d string) (int, bool) {
	t, err := time.Parse("20060102", strings.TrimSpace(d))
	if err != nil {
		return 0, false
	}
	return int(t.Unix() / 86400), true
}

func dayStr(n int) string { return time.Unix(int64(n)*86400, 0).UTC().Format("2006-01-02") }

func isoDate(d string) string {
	if len(d) == 8 {
		return d[:4] + "-" + d[4:6] + "-" + d[6:]
	}
	return d
}

// weekday returns 0 for Sunday; 1970-01-01 was a Thursday.
func weekday(d int) int { return (d + 4) % 7 }

var weekdayNames = [7]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

func daysOfWeek(on [7]bool) string {
	var out []string
	for _, w := range []int{1, 2, 3, 4, 5, 6, 0} {
		if on[w] {
			out = append(out, weekdayNames[w])
		}
	}
	return strings.Join(out, " ")
}

type calException struct {
	day       int
	available bool
}

// calService is one GTFS service as a weekday pattern over a period plus dated exceptions.
type calService struct {
	id       string
	weekdays [7]bool
	hasRange bool
	lo, hi   int
	bits     []byte // exact service days from lo to hi, for UicOperatingPeriod
	except   []calException
}

type calRow struct {
	sid  string
	wd   [7]bool
	a, b int
	ok   bool
}

// dateRow is one calendar_dates row; typ is '1' (added) or '2' (removed).
type dateRow struct {
	day int32
	typ byte
}

// buildCalendar keeps calendar.txt services as written and turns calendar_dates-only services into a majority weekday pattern when that is shorter.
func buildCalendar(cals []calRow, dates map[string][]dateRow, tripServices []string) ([]calService, [2]int) {
	calOf := map[string]*calRow{}
	for i := range cals {
		calOf[cals[i].sid] = &cals[i]
	}
	seen := map[string]bool{}
	var ids []string
	add := func(s string) {
		if !seen[s] {
			seen[s] = true
			ids = append(ids, s)
		}
	}
	for s := range calOf {
		add(s)
	}
	for s := range dates {
		add(s)
	}
	for _, s := range tripServices {
		add(s)
	}
	sort.Strings(ids)
	rng := [2]int{1 << 31, -1}
	res := make([]calService, 0, len(ids))
	days := make([]int, 0, 1024)
	for _, s := range ids {
		cs := calService{id: s}
		ds := dates[s]
		if c := calOf[s]; c != nil {
			active := map[int]bool{}
			if c.ok {
				cs.weekdays, cs.hasRange, cs.lo, cs.hi = c.wd, true, c.a, c.b
				for d := c.a; d <= c.b; d++ {
					if c.wd[weekday(d)] {
						active[d] = true
					}
				}
			}
			for _, d := range ds {
				cs.except = append(cs.except, calException{int(d.day), d.typ == '1'})
				if d.typ == '1' {
					active[int(d.day)] = true
				} else {
					delete(active, int(d.day))
				}
			}
			if cs.hasRange {
				cs.bits = make([]byte, 0, cs.hi-cs.lo+1)
				for d := cs.lo; d <= cs.hi; d++ {
					cs.bits = append(cs.bits, bit(active[d]))
				}
			}
			for d := range active {
				rng[0], rng[1] = min(rng[0], d), max(rng[1], d)
			}
			res = append(res, cs)
			continue
		}
		// calendar_dates-only: active days are the added dates, sorted and unique, minus removed dates.
		days = days[:0]
		var removed map[int]bool
		for _, d := range ds {
			if d.typ == '1' {
				days = append(days, int(d.day))
			} else {
				if removed == nil {
					removed = map[int]bool{}
				}
				removed[int(d.day)] = true
			}
		}
		sort.Ints(days)
		k := 0
		for i, d := range days {
			if (i == 0 || d != days[i-1]) && !removed[d] {
				days[k] = d
				k++
			}
		}
		days = days[:k]
		if len(days) > 0 {
			lo, hi := days[0], days[len(days)-1]
			var total, on [7]int
			for d := lo; d <= hi; d++ {
				total[weekday(d)]++
			}
			for _, d := range days {
				on[weekday(d)]++
			}
			var wd [7]bool
			anyDay := false
			for w := 0; w < 7; w++ {
				wd[w] = on[w]*2 > total[w]
				anyDay = anyDay || wd[w]
			}
			var ex []calException
			bits := make([]byte, 0, hi-lo+1)
			j := 0
			for d := lo; d <= hi; d++ {
				isOn := j < len(days) && days[j] == d
				if isOn {
					j++
				}
				bits = append(bits, bit(isOn))
				if wd[weekday(d)] != isOn {
					ex = append(ex, calException{d, isOn})
				}
			}
			if anyDay && 1+len(ex) < len(days) {
				cs.weekdays, cs.hasRange, cs.lo, cs.hi, cs.except, cs.bits = wd, true, lo, hi, ex, bits
			}
		}
		if !cs.hasRange {
			for _, d := range days {
				cs.except = append(cs.except, calException{d, true})
			}
		}
		if len(days) > 0 {
			rng[0], rng[1] = min(rng[0], days[0]), max(rng[1], days[len(days)-1])
		}
		res = append(res, cs)
	}
	return res, rng
}

func bit(on bool) byte {
	if on {
		return '1'
	}
	return '0'
}
