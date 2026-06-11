package panel

import (
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func moduleAccountBalance(d model.Report, name string) string {
	for _, m := range d.ModuleAccounts {
		if m.Name == name {
			return m.Balance
		}
	}
	return ""
}

func moduleAccountAddress(d model.Report, name string) string {
	for _, m := range d.ModuleAccounts {
		if m.Name == name {
			return m.Address
		}
	}
	return ""
}

// displayAddress prefers EVM hex (0x…) for module account addresses; falls back to bech32.
func displayAddress(addr string) string {
	if addr == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(addr), "0x") {
		return addr
	}
	if evm := fetch.AccBech32ToEVM(addr); evm != "" {
		return evm
	}
	return addr
}

func moduleAccountDisplayAddress(d model.Report, name string) string {
	return displayAddress(moduleAccountAddress(d, name))
}

func writeModuleAccountRow(b *strings.Builder, d model.Report, name, effect string) {
	bal := moduleAccountBalance(d, name)
	addr := moduleAccountDisplayAddress(d, name)
	if bal == "" && addr == "" {
		return
	}
	val := ecoBalanceAddrHTML(orEcoDash(bal), addr)
	ecoDomainRowHTML(b, "", name, val, effect)
}
