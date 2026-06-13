package fetch

import "testing"

func TestChainRecipeNodeMinimalScope(t *testing.T) {
	r := ChainRecipeNode
	if r.ValidatorRewards || r.Governance || r.MintData || r.LocalStaking {
		t.Fatalf("node recipe should be minimal: %+v", r)
	}
	if r.ValidatorScope != ValidatorsBonded {
		t.Fatalf("node should only query bonded validators, got %v", r.ValidatorScope)
	}
}

func TestChainRecipeFullIncludesEverything(t *testing.T) {
	r := ChainRecipeFull
	if !r.CometExtended || !r.Governance || !r.ValidatorRewards || !r.LocalStaking {
		t.Fatalf("full recipe missing sections: %+v", r)
	}
	if r.ValidatorScope != ValidatorsAll {
		t.Fatalf("full recipe should query all validator statuses")
	}
}

func TestFetchChainRecipeStatusRequired(t *testing.T) {
	snap := FetchChainRecipe("http://127.0.0.1:1", "http://127.0.0.1:1", ChainRecipeGovernance)
	if snap.Err == nil {
		t.Fatal("expected error from unreachable RPC")
	}
}
