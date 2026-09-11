package converter

import (
	"testing"

	"github.com/theoremus-urban-solutions/gtfs-netex-converter/gtfs"
)

func TestOrganisationIndexUsesAgencyData(t *testing.T) {
	idx := NewOrganisationIndex("BG", []gtfs.Agency{
		{AgencyID: "BDZ", AgencyName: "Bulgarian State Railways", AgencyURL: "https://bdz.bg", AgencyPhone: "+359 2 931 1111"},
		{AgencyID: "SOF", AgencyName: "Sofia Urban Mobility", AgencyEmail: "info@sofiatraffic.bg"},
	})

	if got := idx.Primary().AuthorityID; got != "BG::Authority:authority_BDZ::" {
		t.Fatalf("primary authority = %q", got)
	}
	if got := idx.ForAgency("SOF").OperatorID; got != "BG::Operator:operator_SOF::" {
		t.Fatalf("SOF operator = %q", got)
	}
	if got := idx.ForAgency("missing").OperatorID; got != idx.Primary().OperatorID {
		t.Fatalf("unknown agency should fall back to primary, got %q", got)
	}
	if got := idx.FrameID("CompositeFrame"); got != "BG::CompositeFrame:BDZ::" {
		t.Fatalf("frame id = %q", got)
	}

	orgs := idx.Organisations()
	if len(orgs.Authorities) != 2 || len(orgs.Operators) != 2 {
		t.Fatalf("expected 2 authorities and 2 operators, got %d/%d", len(orgs.Authorities), len(orgs.Operators))
	}
	if orgs.Operators[0].Name != "Bulgarian State Railways" || orgs.Operators[0].ContactDetails.Url != "https://bdz.bg" {
		t.Fatalf("operator not built from agency.txt: %+v", orgs.Operators[0])
	}
	if orgs.Authorities[1].ContactDetails.Email != "info@sofiatraffic.bg" {
		t.Fatalf("authority contact not built from agency.txt: %+v", orgs.Authorities[1])
	}
	if idx.DataSource().Name != "Bulgarian State Railways" {
		t.Fatalf("data source name = %q", idx.DataSource().Name)
	}
}

func TestOrganisationIndexFallbacks(t *testing.T) {
	idx := NewOrganisationIndex("BG", nil)
	if got := idx.Primary().OperatorID; got != "BG::Operator:operator_UNKNOWN::" {
		t.Fatalf("empty feed operator = %q", got)
	}

	idx = NewOrganisationIndex("BG", []gtfs.Agency{{AgencyName: "Градски Транспорт: Варна"}})
	if got := idx.Primary().Key; got != "Градски_Транспорт_Варна" {
		t.Fatalf("key from name = %q", got)
	}
}
