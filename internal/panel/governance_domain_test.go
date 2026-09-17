package panel

import (
	"strings"
	"testing"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func governanceChunk(t *testing.T, out string) string {
	t.Helper()
	idx := strings.Index(out, `class="dash-heading">6. GOVERNANCE</h2>`)
	if idx < 0 {
		t.Fatal("expected governance section")
	}
	end := strings.Index(out, `class="dash-heading">1. INFRASTRUCTURE</h2>`)
	if end < 0 {
		end = len(out)
	}
	return out[idx:end]
}

func TestGovernanceDomainCards(t *testing.T) {
	d := model.Report{
		VotingPeriod: "2 weeks", ExpeditedVotingPeriod: "1d 0h",
		Quorum: 33.4, Threshold: 50, VetoThreshold: 33.4,
		Proposals:        []model.Proposal{{ID: 1, Title: "test"}},
		DepositProposals: []model.Proposal{{ID: 2}},
		UpgradeName:      "v2", UpgradeHeight: "1000", BlocksLeft: "500",
		IBCClients: 3, ERC20Enabled: true,
		TokenPairs: []model.TokenPair{{Denom: "apmt", ERC20: "0xabc", Enabled: true}},
		ModuleAccounts: []model.ModuleAccountRow{
			{Name: "gov", Balance: "100 PMT", Address: "cosmos1gov"},
		},
	}
	chunk := governanceChunk(t, Build(d))
	for _, want := range []string{
		`eco-domain--gov`,
		`eco-domain--upgrade`,
		`eco-domain--ibc`,
		`eco-domain--erc20`,
		`class="eco-acct__addr">0xabc`,
		"enable_erc20",
		"gov",
		"expedited voting period",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("governance missing %q", want)
		}
	}
	for _, gone := range []string{"Voting Params", "Subsection(\"Upgrade\")"} {
		if strings.Contains(chunk, gone) {
			t.Fatalf("governance should not contain flat %q subsection", gone)
		}
	}
}

func TestGovernanceShowsProposalBody(t *testing.T) {
	d := model.Report{
		Proposals: []model.Proposal{{
			ID: 2, Title: "Set expedited voting period to 1 hour",
			Summary:  "Copy live x/gov params and set only expedited_voting_period from 24h to 1h.",
			Messages: "MsgUpdateParams", Expedited: true, End: "2026-09-18 12:16 UTC",
			MessagesJSON: "{\n  \"@type\": \"/cosmos.gov.v1.MsgUpdateParams\",\n  \"params\": {\"expedited_voting_period\": \"3600s\"}\n}",
			TallyYes:     "1", HasTally: true,
		}},
		RecentProposals: []model.Proposal{{
			ID: 1, Title: "Set feemarket min_gas_price to 1 apmt per gas",
			Status: "passed", End: "2026-09-16",
		}},
	}
	chunk := governanceChunk(t, Build(d))
	for _, want := range []string{
		"Set expedited voting period to 1 hour",
		"Copy live x/gov params and set only expedited_voting_period",
		"expedited",
		"MsgUpdateParams",
		"messages JSON",
		"expedited_voting_period",
		"Recent Proposals",
		"Set feemarket min_gas_price to 1 apmt per gas",
	} {
		if !strings.Contains(chunk, want) {
			t.Fatalf("governance missing %q", want)
		}
	}
}
