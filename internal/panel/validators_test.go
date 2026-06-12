package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestValidatorsP2PNetworkTable(t *testing.T) {
	d := model.Report{
		Validators: []model.Validator{{
			Moniker:         "node1",
			Operator:        "cosmosvaloper1abc",
			P2PDial:         "7c90c689@host:26656",
			NodeID:          "7c90c689deadbeef",
			ConsensusBech32: "cosmosvalcons1xyz",
			IsLocal:         true,
		}},
	}
	chunk := validatorsChunk(t, BuildView(ViewNode, d))

	for _, want := range []string{
		`class="dash-subheading">Network (P2P)</h3>`,
		`data-table--identity`,
		`<th>operator</th>`,
		`<th>p2p dial</th>`,
		`<code>cosmosvaloper1abc</code>`,
		`<code>7c90c689@host:26656</code>`,
		`class="data-table__row--local" title="this node"`,
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("validators P2P section missing %q\n%s", want, chunk)
		}
	}
	for _, bad := range []string{
		`<th>local</th>`,
		`class="kpi-tile kpi-tile--hash"`,
		`class="validator-label"`,
		`class="kpi-tile__primary"`,
	} {
		if strings.Contains(chunk, bad) {
			t.Fatalf("validators P2P section should not use KPI tiles, found %q", bad)
		}
	}
}

func validatorsChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">2. VALIDATOR</h2>`)
	if idx < 0 {
		t.Fatal("expected validator section")
	}
	return out[idx:]
}
