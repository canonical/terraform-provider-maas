package maas

import (
	"testing"

	"github.com/canonical/gomaasclient/entity"
)

// A fabric always holds an untagged VLAN, and its database ID can collide with
// another VLAN's VID. MAAS returns VLANs in PK order with no ordering clause,
// so the winner must come from the lookup mode, not from the slice order.
func TestMatchVLANResolvesIdentifiersDeterministically(t *testing.T) {
	untagged := entity.VLAN{ID: 20, VID: 0, FabricID: 7}
	tagged := entity.VLAN{ID: 21, VID: 20, FabricID: 7}
	other := entity.VLAN{ID: 22, VID: 5, FabricID: 7}

	ordered := []entity.VLAN{untagged, tagged, other}
	reversed := []entity.VLAN{other, tagged, untagged}

	tests := []struct {
		name       string
		identifier string
		lookup     vlanLookup
		wantID     int // zero means no match
	}{
		{"VID wins when it matches an ID as well", "20", lookupVLANByVID, tagged.ID},
		{"ID wins for fields documented as an ID", "20", lookupVLANByID, untagged.ID},
		{"VID lookup falls back to the ID", "22", lookupVLANByVID, other.ID},
		{"ID lookup falls back to the VID", "5", lookupVLANByID, other.ID},
		{"untagged VLAN is found by VID", "0", lookupVLANByVID, untagged.ID},
		{"unknown identifier matches nothing", "99", lookupVLANByVID, 0},
	}

	for _, tt := range tests {
		for name, order := range map[string][]entity.VLAN{"as returned": ordered, "reversed": reversed} {
			t.Run(tt.name+"/"+name, func(t *testing.T) {
				got := matchVLAN(order, tt.identifier, tt.lookup)

				if tt.wantID == 0 {
					if got != nil {
						t.Fatalf("matchVLAN(%q) = VLAN %d, want no match", tt.identifier, got.ID)
					}

					return
				}

				if got == nil {
					t.Fatalf("matchVLAN(%q) found nothing, want VLAN %d", tt.identifier, tt.wantID)
				}

				if got.ID != tt.wantID {
					t.Errorf("matchVLAN(%q) = VLAN %d (VID %d), want VLAN %d",
						tt.identifier, got.ID, got.VID, tt.wantID)
				}
			})
		}
	}
}
