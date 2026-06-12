package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestStatusStripUnknownPeersAndNode(t *testing.T) {
	out := RenderStatusStrip(model.Report{
		TimeUTC: "12:00:00 UTC",
	})
	if !strings.Contains(out, ">Peers</span><span class=\"dash-status__value\">—</span>") {
		t.Fatalf("expected unknown peers dash, got: %s", out)
	}
	if !strings.Contains(out, ">Node</span><span class=\"dash-status__value\">—</span>") {
		t.Fatalf("expected unknown node dash, got: %s", out)
	}
	if strings.Contains(out, "badge--bad") {
		t.Fatalf("should not show stopped badge when node status unknown: %s", out)
	}
}

func TestStatusStripLiveValues(t *testing.T) {
	out := RenderStatusStrip(model.Report{
		HasChainStatus: true,
		HasEVMPeers:    true,
		HasNodeStatus:  true,
		Synced:         true,
		BlockHeight:    "100",
		PeerCount:      3,
		EVMPeerCount:   1,
		NodeRunning:    true,
		BaseFee:        "7 apmt",
		PMTEnabled:     true,
		TimeUTC:        "12:00:00 UTC",
	})
	for _, want := range []string{
		">100<",
		">synced<",
		">Peers</span><span class=\"dash-status__value\">3</span>",
		">running<",
		"7 apmt",
		">enabled<",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in: %s", want, out)
		}
	}
}

func TestStatusStripPartialPeers(t *testing.T) {
	out := RenderStatusStrip(model.Report{
		HasChainStatus: true,
		PeerCount:      5,
		BlockHeight:    "1",
		Synced:         true,
		TimeUTC:        "12:00:00 UTC",
	})
	if !strings.Contains(out, ">Peers</span><span class=\"dash-status__value\">5</span>") {
		t.Fatalf("expected cosmos peer count only, got: %s", out)
	}
}
