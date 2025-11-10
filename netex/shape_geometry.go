package netex

// PointOnLink represents a point on a geographic link
type PointOnLink struct {
	ID                 string   `xml:"id,attr"`
	Version            string   `xml:"version,attr"`
	Order              int      `xml:"order,attr"`
	Location           Location `xml:"Location"`
	DistanceFromStart  *float64 `xml:"DistanceFromStart,omitempty"`
}

// PointsOnLink represents a collection of points on a link
type PointsOnLink struct {
	PointOnLink []PointOnLink `xml:"PointOnLink"`
}

// LinkSequence represents a sequence of route links
type LinkSequence struct {
	ID        string            `xml:"id,attr"`
	Version   string            `xml:"version,attr"`
	LinkRefs  LinkRefs          `xml:"links"`
}

// LinkRefs represents references to route links
type LinkRefs struct {
	RouteLinkRef []RouteLinkRef `xml:"RouteLinkRef"`
}

// RouteLinkRef represents a reference to a route link
type RouteLinkRef struct {
	Ref     string `xml:"ref,attr"`
	Version string `xml:"version,attr"`
}