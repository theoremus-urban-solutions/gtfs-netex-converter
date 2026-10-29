package stream

import (
	"bufio"
	"io"
	"strconv"
	"unicode/utf8"
)

// out writes XML text straight to the destination, escaping as encoding/xml does.
type out struct {
	w       *bufio.Writer
	n       int64
	scratch []byte
}

func newOut(w io.Writer) *out {
	return &out{w: bufio.NewWriterSize(w, 1<<16), scratch: make([]byte, 0, 1<<12)}
}

func (o *out) s(x string) {
	o.n += int64(len(x))
	_, _ = o.w.WriteString(x)
}

func (o *out) b(x []byte) {
	o.n += int64(len(x))
	_, _ = o.w.Write(x)
}

// escape also replaces invalid UTF-8 and characters XML 1.0 forbids with U+FFFD, as encoding/xml does.
func escape(b []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			r, n := utf8.DecodeRuneInString(s[i:])
			if (r == utf8.RuneError && n == 1) || r == 0xFFFE || r == 0xFFFF {
				b = append(b, "\uFFFD"...)
			} else {
				b = append(b, s[i:i+n]...)
			}
			i += n - 1
			continue
		}
		switch c := s[i]; c {
		case '"':
			b = append(b, "&#34;"...)
		case '\'':
			b = append(b, "&#39;"...)
		case '&':
			b = append(b, "&amp;"...)
		case '<':
			b = append(b, "&lt;"...)
		case '>':
			b = append(b, "&gt;"...)
		case '\t':
			b = append(b, "&#x9;"...)
		case '\n':
			b = append(b, "&#xA;"...)
		case '\r':
			b = append(b, "&#xD;"...)
		default:
			if c < 0x20 {
				b = append(b, "\uFFFD"...)
			} else {
				b = append(b, c)
			}
		}
	}
	return b
}

// e writes escaped text.
func (o *out) e(x string) {
	o.scratch = escape(o.scratch[:0], x)
	o.b(o.scratch)
}

func (o *out) el(tag, text string) {
	o.s("<" + tag + ">")
	o.e(text)
	o.s("</" + tag + ">")
}

func (o *out) elOpt(tag, text string) {
	if text != "" {
		o.el(tag, text)
	}
}

// ref writes <tag ref="r" version="1"></tag>.
func (o *out) ref(tag, r string) {
	o.s("<" + tag + ` ref="`)
	o.e(r)
	o.s(`" version="1"></` + tag + ">")
}

func (o *out) open(tag, id string) {
	o.s("<" + tag + ` id="`)
	o.e(id)
	o.s(`" version="1">`)
}

// frameType writes an EPIP TypeOfFrameRef such as epip:EU_PI_NETWORK; versionRef, not version, so no TypeOfFrame must be defined in the file.
func (o *out) frameType(name string) {
	o.s(`<TypeOfFrameRef ref="epip:` + name + `" versionRef="epip:1.0"></TypeOfFrameRef>`)
}

func gf(f float64) string { return strconv.FormatFloat(f, 'g', -1, 64) }

func (o *out) loc(lon, lat float64) {
	o.s("<Location><Longitude>" + gf(lon) + "</Longitude><Latitude>" + gf(lat) + "</Latitude></Location>")
}

// flush returns the first write error seen.
func (o *out) flush() error { return o.w.Flush() }
