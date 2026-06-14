package fetch

import (
	"fmt"
	"sync"
)

// ModuleBalanceInfo is a Cosmos SDK module account bank balance.
type ModuleBalanceInfo struct {
	Name    string
	Address string
	Amount  string
	Denom   string
}

var trackedModuleAccounts = []struct {
	name string
	role string
}{
	{"fee_collector", "Fees + minted rewards land here each block, then distribution clears"},
	{"distribution", "x/distribution module escrow (often ~0 after BeginBlock payout)"},
	{"bonded_tokens_pool", "Staked tokens (locked; matches staking pool bonded)"},
	{"not_bonded_tokens_pool", "Unbonding / unbonded stake in staking pool"},
	{"gov", "Proposal deposit escrow until voting or refund"},
}

type moduleAccountsResp struct {
	Accounts []struct {
		Name        string `json:"name"`
		BaseAccount struct {
			Address string `json:"address"`
		} `json:"base_account"`
	} `json:"accounts"`
}

type bankBalancesResp struct {
	Balances []struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	} `json:"balances"`
}

// FetchModuleBalances returns live bank balances for core economics module accounts.
// When names is non-empty, only those module account names are queried.
func FetchModuleBalances(rest, preferDenom string, names []string) []ModuleBalanceInfo {
	specs := trackedModuleAccounts
	if len(names) > 0 {
		want := make(map[string]bool, len(names))
		for _, n := range names {
			want[n] = true
		}
		filtered := make([]struct {
			name string
			role string
		}, 0, len(names))
		for _, s := range trackedModuleAccounts {
			if want[s.name] {
				filtered = append(filtered, s)
			}
		}
		specs = filtered
	}

	var ma moduleAccountsResp
	if err := doJSON(rest+"/cosmos/auth/v1beta1/module_accounts", &ma); err != nil {
		return nil
	}
	addrByName := make(map[string]string, len(ma.Accounts))
	for _, acct := range ma.Accounts {
		if acct.Name != "" && acct.BaseAccount.Address != "" {
			addrByName[acct.Name] = acct.BaseAccount.Address
		}
	}

	out := make([]ModuleBalanceInfo, len(specs))
	var wg sync.WaitGroup
	for i, spec := range specs {
		i, spec := i, spec
		out[i].Name = spec.name
		addr, ok := addrByName[spec.name]
		if !ok {
			continue
		}
		out[i].Address = addr
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()
			var bal bankBalancesResp
			if err := doJSON(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", rest, address), &bal); err != nil {
				return
			}
			for _, b := range bal.Balances {
				if preferDenom != "" && b.Denom != preferDenom {
					continue
				}
				out[idx].Amount = b.Amount
				out[idx].Denom = b.Denom
				return
			}
			if len(bal.Balances) > 0 {
				out[idx].Amount = bal.Balances[0].Amount
				out[idx].Denom = bal.Balances[0].Denom
			}
		}(i, addr)
	}
	wg.Wait()
	return out
}
