package fetch

import (
	"fmt"
	"net/url"
	"time"
)

// DelegationInfo is a single delegator entry for a validator.
type DelegationInfo struct {
	DelegatorAddr      string
	BalanceAmt         string
	BalanceDenom       string
	Shares             string
	LiquidBalanceAmt   string
	LiquidBalanceDenom string
}

// UnbondingEntryInfo is one unbonding entry for a delegator→validator pair.
type UnbondingEntryInfo struct {
	Balance        string
	CompletionTime time.Time
}

// UnbondingDelegationInfo is unbonding stake from one delegator to a validator.
type UnbondingDelegationInfo struct {
	DelegatorAddr string
	Entries       []UnbondingEntryInfo
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

type unbondingDelegationsResp struct {
	UnbondingResponses []struct {
		DelegatorAddress string `json:"delegator_address"`
		Entries          []struct {
			Balance        string `json:"balance"`
			CompletionTime string `json:"completion_time"`
		} `json:"entries"`
	} `json:"unbonding_responses"`
	Pagination struct {
		NextKey string `json:"next_key"`
	} `json:"pagination"`
}

// FetchValidatorUnbondingDelegations returns unbonding entries targeting a validator.
func FetchValidatorUnbondingDelegations(rest, valoper string) []UnbondingDelegationInfo {
	if valoper == "" {
		return nil
	}
	var out []UnbondingDelegationInfo
	nextKey := ""
	for page := 0; page < 32; page++ {
		reqURL := fmt.Sprintf("%s/cosmos/staking/v1beta1/validators/%s/unbonding_delegations?pagination.limit=100",
			rest, url.PathEscape(valoper))
		if nextKey != "" {
			reqURL += "&pagination.key=" + url.QueryEscape(nextKey)
		}
		var resp unbondingDelegationsResp
		if err := doJSON(reqURL, &resp); err != nil {
			break
		}
		for _, u := range resp.UnbondingResponses {
			info := UnbondingDelegationInfo{DelegatorAddr: u.DelegatorAddress}
			for _, e := range u.Entries {
				entry := UnbondingEntryInfo{Balance: e.Balance}
				if t, err := time.Parse(time.RFC3339Nano, e.CompletionTime); err == nil {
					entry.CompletionTime = t
				} else if t, err := time.Parse(time.RFC3339, e.CompletionTime); err == nil {
					entry.CompletionTime = t
				}
				info.Entries = append(info.Entries, entry)
			}
			if len(info.Entries) > 0 {
				out = append(out, info)
			}
		}
		nextKey = resp.Pagination.NextKey
		if nextKey == "" {
			break
		}
	}
	return out
}

// SummarizeUnbondings returns total raw balance, entry count, and earliest completion.
func SummarizeUnbondings(unbondings []UnbondingDelegationInfo) (totalRaw string, entries int, earliest time.Time) {
	amts := make([]string, 0, 8)
	for _, u := range unbondings {
		for _, e := range u.Entries {
			if e.Balance == "" || e.Balance == "0" {
				continue
			}
			amts = append(amts, e.Balance)
			entries++
			if !e.CompletionTime.IsZero() && (earliest.IsZero() || e.CompletionTime.Before(earliest)) {
				earliest = e.CompletionTime
			}
		}
	}
	if len(amts) == 0 {
		return "", 0, time.Time{}
	}
	return SumRawAmounts(amts...), entries, earliest
}
