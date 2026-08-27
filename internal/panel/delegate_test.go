package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func TestParseViewDelegate(t *testing.T) {
	if ParseView("delegate") != ViewDelegate {
		t.Fatal("ParseView(delegate) should be ViewDelegate")
	}
}

func TestBuildViewDelegate(t *testing.T) {
	out := BuildViewWithOptions(ViewDelegate, model.Report{Moniker: "node1"}, Options{PublicEVM: "https://node1.pmtchain.com"})
	for _, want := range []string{
		StakingPrecompile,
		"290290",
		"https://node1.pmtchain.com",
		"id=\"delegate-app\"",
		"Connect MetaMask",
		"cosmosvaloper1akkvh0ahmve830rj4mhkdnqs49kzw23cl98zp4",
		"cosmosvaloper1r2dqta25pxj8av9grlfxvnfje006papu0tjk0a",
		"cosmosvaloper15hr4x4rfj0y82puk74xegugn5s5clphzcfej3e",
		"cosmosvaloper1vmr9wxpldngnh0tvpr8h2pk2aycts3v7z8pdxh",
		"Another validator",
		"id=\"delegate-picker\"",
		"id=\"delegate-valoper\"",
		"disabled",
		"id=\"delegate-error\"",
		"id=\"delegate-error-submit\"",
		"Leave liquid PMT for gas",
		"Pick node1",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("delegate view missing %q", want)
		}
	}
	for _, bad := range []string{
		"Our four",
		"we run the nodes",
		"Other — paste valoper",
		"id=\"delegate-preset-wrap\"",
		"id=\"delegate-custom-wrap\"",
		"delegate-preset-valoper",
	} {
		if strings.Contains(out, bad) {
			t.Fatalf("delegate view should not contain %q", bad)
		}
	}
	if strings.Contains(out, "hx-trigger") {
		t.Fatal("delegate fragment should not poll")
	}
}

func TestStakingLinksToDelegate(t *testing.T) {
	out := BuildView(ViewStaking, model.Report{Moniker: "n", Synced: true, BlockHeight: "1"})
	if !strings.Contains(out, `href="/delegate"`) || !strings.Contains(out, `hx-boost="false"`) {
		t.Fatal("staking monitor should full-load /delegate so wallet scripts run")
	}
}
