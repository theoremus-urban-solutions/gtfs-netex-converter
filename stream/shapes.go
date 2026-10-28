package stream

import (
	"math"
	"sort"
	"strconv"
)

// maxStopOffset is how far, in metres, a stop may lie from its shape and still cut it.
const maxStopOffset = 1000.0

// maxCandidates bounds the places on a shape that one stop may cut it.
const maxCandidates = 8

type shapeCandidate struct {
	pos  float64 // segment index plus fraction along it
	dist float64
}

// cutShape splits a shape of lat, lon pairs at the stops' points, in stop order.
// It returns one lat, lon slice per consecutive stop pair, or nil when the stops cannot be placed on the shape in order.
func cutShape(shape []float64, stops [][2]float64) [][]float64 {
	n := len(shape) / 2
	if n < 2 || len(stops) < 2 {
		return nil
	}
	lat0 := shape[0] * math.Pi / 180
	kx, ky := 111320*math.Cos(lat0), 110540.0
	px := make([]float64, n)
	py := make([]float64, n)
	for i := 0; i < n; i++ {
		px[i], py[i] = shape[2*i+1]*kx, shape[2*i]*ky
	}
	dist := make([]float64, n-1)
	frac := make([]float64, n-1)
	cands := make([][]shapeCandidate, len(stops))
	for s, st := range stops {
		x, y := st[1]*kx, st[0]*ky
		for k := 0; k < n-1; k++ {
			dx, dy := px[k+1]-px[k], py[k+1]-py[k]
			t := 0.0
			if l2 := dx*dx + dy*dy; l2 > 0 {
				t = math.Max(0, math.Min(1, ((x-px[k])*dx+(y-py[k])*dy)/l2))
			}
			frac[k] = t
			dist[k] = math.Hypot(px[k]+t*dx-x, py[k]+t*dy-y)
		}
		var cs []shapeCandidate
		for k := 0; k < n-1; k++ {
			if dist[k] > maxStopOffset {
				continue
			}
			if (k > 0 && dist[k-1] < dist[k]) || (k < n-2 && dist[k+1] < dist[k]) {
				continue
			}
			cs = append(cs, shapeCandidate{float64(k) + frac[k], dist[k]})
		}
		if len(cs) == 0 {
			return nil
		}
		if len(cs) > maxCandidates {
			sort.Slice(cs, func(a, b int) bool { return cs[a].dist < cs[b].dist })
			cs = cs[:maxCandidates]
			sort.Slice(cs, func(a, b int) bool { return cs[a].pos < cs[b].pos })
		}
		cands[s] = cs
	}
	// Pick one candidate per stop, never moving backwards, with the least total offset.
	cost := make([][]float64, len(stops))
	back := make([][]int, len(stops))
	cost[0] = make([]float64, len(cands[0]))
	for c, cd := range cands[0] {
		cost[0][c] = cd.dist
	}
	for s := 1; s < len(stops); s++ {
		cost[s] = make([]float64, len(cands[s]))
		back[s] = make([]int, len(cands[s]))
		for c, cd := range cands[s] {
			cost[s][c], back[s][c] = math.Inf(1), -1
			for p, pd := range cands[s-1] {
				if pd.pos <= cd.pos && cost[s-1][p]+cd.dist < cost[s][c] {
					cost[s][c], back[s][c] = cost[s-1][p]+cd.dist, p
				}
			}
		}
	}
	last := len(stops) - 1
	best := -1
	for c := range cands[last] {
		if !math.IsInf(cost[last][c], 1) && (best < 0 || cost[last][c] < cost[last][best]) {
			best = c
		}
	}
	if best < 0 {
		return nil
	}
	pos := make([]float64, len(stops))
	for s := last; s >= 0; s-- {
		pos[s] = cands[s][best].pos
		if s > 0 {
			best = back[s][best]
		}
	}
	at := func(p float64) (float64, float64) {
		k := int(p)
		if k >= n-1 {
			return shape[2*(n-1)], shape[2*(n-1)+1]
		}
		t := p - float64(k)
		return round6(shape[2*k] + t*(shape[2*k+2]-shape[2*k])), round6(shape[2*k+1] + t*(shape[2*k+3]-shape[2*k+1]))
	}
	pieces := make([][]float64, 0, last)
	for s := 0; s < last; s++ {
		a, b := pos[s], pos[s+1]
		la, lo := at(a)
		piece := []float64{la, lo}
		for v := int(a) + 1; float64(v) < b; v++ {
			piece = append(piece, shape[2*v], shape[2*v+1])
		}
		la, lo = at(b)
		piece = append(piece, la, lo)
		pieces = append(pieces, piece)
	}
	return pieces
}

func round6(v float64) float64 { return math.Round(v*1e6) / 1e6 }

// appendPosList writes lat, lon pairs as a gml:posList value.
func appendPosList(b []byte, pts []float64) []byte {
	for i, v := range pts {
		if i > 0 {
			b = append(b, ' ')
		}
		b = strconv.AppendFloat(b, v, 'f', -1, 64)
	}
	return b
}
