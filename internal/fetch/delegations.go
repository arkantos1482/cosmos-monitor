package fetch

import (
	"fmt"
	"net/url"
)

// DelegationInfo is a single delegator entry for a validator.
type DelegationInfo struct {
	DelegatorAddr    string
	BalanceAmt       string
	BalanceDenom     string
	Shares           string
	LiquidBalanceAmt string
	LiquidBalanceDenom string
}

type delegationsResp struct {
	DelegationResponses []struct {
		Delegation struct {
			DelegatorAddress string `json:"delegator_address"`
			ValidatorAddress string `json:"validator_address"`
			Shares           string `json:"shares"`
		} `json:"delegation"`
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		} `json:"balance"`
	} `json:"delegation_responses"`
	Pagination struct {
		NextKey string `json:"next_key"`
	} `json:"pagination"`
}

// FetchValidatorDelegations returns all delegations for a validator operator address.
func FetchValidatorDelegations(rest, valoper string) []DelegationInfo {
	if valoper == "" {
		return nil
	}
	var out []DelegationInfo
	nextKey := ""
	for page := 0; page < 32; page++ {
		reqURL := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/delegations?pagination.limit=100",
			rest, url.PathEscape(valoper))
		if nextKey != "" {
			reqURL += "&pagination.key=" + url.QueryEscape(nextKey)
		}
		var resp delegationsResp
		if err := doJSON(reqURL, &resp); err != nil {
			break
		}
		for _, d := range resp.DelegationResponses {
			del := d.Delegation.DelegatorAddress
			if del == "" {
				continue
			}
			out = append(out, DelegationInfo{
				DelegatorAddr: del,
				BalanceAmt:    d.Balance.Amount,
				BalanceDenom:  d.Balance.Denom,
				Shares:        d.Delegation.Shares,
			})
		}
		nextKey = resp.Pagination.NextKey
		if nextKey == "" {
			break
		}
	}
	return out
}

// EnrichDelegationLiquidBalances fills LiquidBalance* from bank balances (preferDenom when set).
func EnrichDelegationLiquidBalances(rest string, delegations []DelegationInfo, preferDenom string) {
	for i := range delegations {
		amt, denom := FetchAddressBalance(rest, delegations[i].DelegatorAddr, preferDenom)
		delegations[i].LiquidBalanceAmt = amt
		delegations[i].LiquidBalanceDenom = denom
	}
}
