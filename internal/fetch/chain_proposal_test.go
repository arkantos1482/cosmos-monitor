package fetch

import (
	"encoding/json"
	"testing"
)

func TestParseProposalV1TitleAndSummary(t *testing.T) {
	raw := []byte(`{
		"id": "2",
		"title": "Set expedited voting period to 1 hour",
		"summary": "Copy live x/gov params and set only expedited_voting_period from 24h to 1h.",
		"status": "PROPOSAL_STATUS_VOTING_PERIOD",
		"expedited": true,
		"messages": [{"@type": "/cosmos.gov.v1.MsgUpdateParams"}],
		"voting_end_time": "2026-09-18T12:16:48.542644779Z"
	}`)
	var p rawProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	got := parseProposal(p)
	if got.ID != 2 || got.Title == "" || got.Summary == "" {
		t.Fatalf("expected title and summary, got %+v", got)
	}
	if !got.Expedited {
		t.Fatal("expected expedited")
	}
	if got.Messages != "MsgUpdateParams" {
		t.Fatalf("messages=%q", got.Messages)
	}
	if proposalStatusKind(got.Status) != "voting" {
		t.Fatalf("kind=%s", proposalStatusKind(got.Status))
	}
}

func TestParseProposalV1Beta1FallsBackToType(t *testing.T) {
	raw := []byte(`{
		"proposal_id": "2",
		"content": {"@type": "/cosmos.gov.v1.MsgUpdateParams", "authority": "cosmos1x"},
		"status": "PROPOSAL_STATUS_VOTING_PERIOD"
	}`)
	var p rawProposal
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatal(err)
	}
	got := parseProposal(p)
	if got.Title != "MsgUpdateParams" {
		t.Fatalf("title=%q want MsgUpdateParams", got.Title)
	}
}

func TestParseTallyAcceptsCountFields(t *testing.T) {
	var tr proposalTallyResp
	if err := json.Unmarshal([]byte(`{"tally":{"yes_count":"10","no_count":"1"}}`), &tr); err != nil {
		t.Fatal(err)
	}
	tally := parseTally(tr)
	if tally.Yes != "10" || tally.No != "1" {
		t.Fatalf("tally=%+v", tally)
	}
}
