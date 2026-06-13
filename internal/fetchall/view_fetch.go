package fetchall

import (
	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
)

func chainRecipeFor(view panel.View) fetch.ChainRecipe {
	switch view {
	case panel.ViewNode:
		return fetch.ChainRecipeNode
	case panel.ViewStaking:
		return fetch.ChainRecipeStaking
	case panel.ViewSlashing:
		return fetch.ChainRecipeSlashing
	case panel.ViewRewards:
		return fetch.ChainRecipeRewards
	case panel.ViewDistribution:
		return fetch.ChainRecipeDistribution
	case panel.ViewFeemarket:
		return fetch.ChainRecipeFeemarket
	case panel.ViewGovernance:
		return fetch.ChainRecipeGovernance
	default:
		return fetch.ChainRecipeFull
	}
}

func needsAppToml(view panel.View) bool {
	return view == panel.ViewHome || view == panel.ViewFeemarket
}

func needsLocalBalanceEnrichment(recipe fetch.ChainRecipe) bool {
	return recipe.LocalStaking
}
