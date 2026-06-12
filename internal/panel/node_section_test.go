package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestNodeSectionConsensusAndP2PInSubsections(t *testing.T) {
	d := model.Report{
		Moniker:  "node1",
		NodeID:   "7C90C68908923B0ABF17CE9BB7D79DD405ABFE95",
		Synced:   true,
		PeerCount: 3,
		Local: model.LocalValidator{
			IsValidator:   true,
			Moniker:       "node1",
			ConsensusAddr: "31aec3d55f45aa21f7efbcc4257ea9f56c9ad300",
			VotingPower:   "100",
		},
	}
	out := BuildView(ViewNode, d)
	proposerEnd := strings.Index(out, `class="dash-subheading">P2P &amp; RPC</h3>`)
	if proposerEnd < 0 {
		t.Fatal("expected P2P subsection")
	}
	proposer := out[:proposerEnd]
	for _, want := range []string{
		`class="dash-subheading">Proposer</h3>`,
		`kpi-tile__label">consensus</div>`,
		`cosmosvalcons`,
	} {
		if !strings.Contains(proposer, want) {
			t.Fatalf("proposer subsection missing %q\n%s", want, proposer)
		}
	}
	p2p := out[proposerEnd:]
	for _, want := range []string{
		`kpi-tile__label">node ID</div>`,
		`7c90c68908923b0abf17ce9bb7d79dd405abfe95`,
		`kpi-tile__label">cosmos peers</div>`,
	} {
		if !strings.Contains(p2p, want) {
			t.Fatalf("P2P subsection missing %q\n%s", want, p2p)
		}
	}
	for _, gone := range []string{
		`class="id-board"`,
		"evm peers",
		`node-summary__label">peers</span><span class="node-summary__val">3 cosmos ·`,
	} {
		if strings.Contains(out, gone) {
			t.Fatalf("node section should not contain %q", gone)
		}
	}
}
