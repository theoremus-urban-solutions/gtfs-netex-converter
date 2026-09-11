package converter

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/gtfs"
	"github.com/theoremus-urban-solutions/gtfs-netex-converter/netex"
)

// fallbackAgencyKey is used when a GTFS feed has no agency.txt rows.
const fallbackAgencyKey = "UNKNOWN"

// OrganisationRefs holds the NeTEx IDs generated for one GTFS agency.
type OrganisationRefs struct {
	Key         string
	AuthorityID string
	OperatorID  string
}

// OrganisationIndex maps GTFS agency IDs to their NeTEx organisation IDs.
type OrganisationIndex struct {
	byAgencyID map[string]OrganisationRefs
	primary    OrganisationRefs
	agencies   []gtfs.Agency
	idGen      *IDGenerator
}

// NewOrganisationIndex builds organisation IDs for every GTFS agency.
// The first agency is the primary one and is used as the default for
// entities that do not carry an agency_id.
func NewOrganisationIndex(participantRef string, agencies []gtfs.Agency) *OrganisationIndex {
	idx := &OrganisationIndex{
		byAgencyID: make(map[string]OrganisationRefs),
		idGen:      NewIDGenerator(participantRef),
	}

	if len(agencies) == 0 {
		agencies = []gtfs.Agency{{
			AgencyID:   fallbackAgencyKey,
			AgencyName: "Unknown operator",
		}}
	}
	idx.agencies = agencies

	for i, agency := range agencies {
		key := agencyKey(agency, i)
		refs := OrganisationRefs{
			Key:         key,
			AuthorityID: idx.idGen.GenerateAuthorityID(key),
			OperatorID:  idx.idGen.GenerateOperatorID(key),
		}
		idx.byAgencyID[agency.AgencyID] = refs
		if i == 0 {
			idx.primary = refs
		}
	}

	return idx
}

// Primary returns the organisation refs of the first agency in the feed.
func (idx *OrganisationIndex) Primary() OrganisationRefs {
	return idx.primary
}

// ForAgency returns the organisation refs for a GTFS agency_id, falling back
// to the primary agency when the ID is empty or unknown.
func (idx *OrganisationIndex) ForAgency(agencyID string) OrganisationRefs {
	if refs, ok := idx.byAgencyID[agencyID]; ok {
		return refs
	}
	return idx.primary
}

// Organisations builds the NeTEx Authority and Operator entities for all agencies.
func (idx *OrganisationIndex) Organisations() netex.Organisations {
	orgs := netex.Organisations{
		Authorities: make([]netex.Authority, 0, len(idx.agencies)),
		Operators:   make([]netex.Operator, 0, len(idx.agencies)),
	}

	for i, agency := range idx.agencies {
		refs := idx.byAgencyID[agency.AgencyID]
		key := agencyKey(agency, i)
		contact := netex.ContactDetails{
			Phone: agency.AgencyPhone,
			Email: agency.AgencyEmail,
			Url:   agency.AgencyURL,
		}

		orgs.Authorities = append(orgs.Authorities, netex.Authority{
			ID:               refs.AuthorityID,
			Version:          "1",
			PublicCode:       key,
			Name:             agency.AgencyName,
			ShortName:        agency.AgencyName,
			LegalName:        agency.AgencyName,
			ContactDetails:   contact,
			OrganisationType: "authority",
		})
		orgs.Operators = append(orgs.Operators, netex.Operator{
			ID:               refs.OperatorID,
			Version:          "1",
			PublicCode:       key,
			Name:             agency.AgencyName,
			ShortName:        agency.AgencyName,
			LegalName:        agency.AgencyName,
			ContactDetails:   contact,
			OrganisationType: "operator",
		})
	}

	return orgs
}

// DataSource builds the DataSource entity for the primary agency.
func (idx *OrganisationIndex) DataSource() netex.DataSource {
	return netex.DataSource{
		ID:      idx.idGen.GenerateID("DataSource", idx.primary.Key),
		Version: "1",
		Name:    idx.agencies[0].AgencyName,
	}
}

// ResponsibilitySet builds the ResponsibilitySet entity for the primary agency.
func (idx *OrganisationIndex) ResponsibilitySet() netex.ResponsibilitySet {
	return netex.ResponsibilitySet{
		ID:      idx.idGen.GenerateID("ResponsibilitySet", idx.primary.Key),
		Version: "1",
		Roles: netex.Roles{
			ResponsibilityRoleAssignment: netex.ResponsibilityRoleAssignment{
				ID:           idx.idGen.GenerateID("ResponsibilityRoleAssignment", idx.primary.Key),
				Version:      "1",
				DataRoleType: "dataSource",
				ResponsibleOrganisationRef: netex.ResponsibleOrganisationRef{
					Ref: idx.primary.AuthorityID,
				},
			},
		},
	}
}

// FrameID builds a frame ID scoped to the primary agency, e.g. BG::CompositeFrame:BDZ::.
func (idx *OrganisationIndex) FrameID(frameType string) string {
	return idx.idGen.GenerateID(frameType, idx.primary.Key)
}

// agencyKey returns an identifier-safe key for an agency. agency_id is optional
// in GTFS single-agency feeds, so fall back to the name, then to the position.
func agencyKey(agency gtfs.Agency, position int) string {
	raw := agency.AgencyID
	if raw == "" {
		raw = agency.AgencyName
	}
	if raw == "" {
		raw = fmt.Sprintf("%s_%d", fallbackAgencyKey, position)
	}

	var b strings.Builder
	for _, r := range raw {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), r == '_', r == '-':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('_')
		default:
			// Drop ':' and other separators that would break NeTEx ID structure.
		}
	}
	if b.Len() == 0 {
		return fmt.Sprintf("%s_%d", fallbackAgencyKey, position)
	}
	return b.String()
}
