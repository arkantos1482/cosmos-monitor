package fetch

import "fmt"

// FetchAddressBalance returns the preferDenom balance for a bech32 account address.
func FetchAddressBalance(rest, addr, preferDenom string) (amount, denom string) {
	if addr == "" {
		return "", ""
	}
	var bal bankBalancesResp
	if err := doJSON(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s", rest, addr), &bal); err != nil {
		return "", ""
	}
	for _, b := range bal.Balances {
		if preferDenom != "" && b.Denom != preferDenom {
			continue
		}
		return b.Amount, b.Denom
	}
	if len(bal.Balances) > 0 {
		return bal.Balances[0].Amount, bal.Balances[0].Denom
	}
	return "", ""
}
