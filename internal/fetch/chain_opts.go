package fetch

// ChainOpts selects optional subsets of FetchChain for lighter dashboard sections.
// Zero value fetches everything.
type ChainOpts struct {
	SkipValidatorRewards bool
	SkipGovernance       bool
	SkipEconomics        bool
}
