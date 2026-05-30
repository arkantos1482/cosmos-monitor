package ui

import (
	"fmt"
)

func renderEVM(s *unifiedSnapshot, width, height int) string {
	if s == nil {
		return panelStyle.Width(width).Height(height).Render(dim("fetching…"))
	}

	evm := s.EVM
	chain := s.Chain
	title := styleBold.Render("EVM")

	if evm.Err != nil {
		return panelStyle.Width(width).Height(height).Render(title + "\n\n" + dim("unavailable: "+evm.Err.Error()))
	}

	lines := []string{title, ""}

	lines = append(lines, fmt.Sprintf("%-12s %d", label("Chain ID"), evm.ChainID))

	var syncStr string
	if evm.Syncing {
		syncStr = bad("syncing ✗")
	} else {
		syncStr = ok("synced ✓")
	}
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Sync"), syncStr))
	lines = append(lines, "")

	// txpool
	var txpoolStr string
	if evm.PendingTx > 50 {
		txpoolStr = warn(fmt.Sprintf("%d pending / %d queued", evm.PendingTx, evm.QueuedTx))
	} else {
		txpoolStr = fmt.Sprintf("%d pending / %d queued", evm.PendingTx, evm.QueuedTx)
	}
	lines = append(lines, fmt.Sprintf("%-12s %s", label("txpool"), txpoolStr))

	// base fee (from chain snapshot, same source)
	baseFee := chain.BaseFee
	if baseFee == "" {
		baseFee = dim("?")
	}
	lines = append(lines, fmt.Sprintf("%-12s %s", label("Base fee"), baseFee))

	// block gas
	if chain.BlockGas > 0 {
		// Cosmos default max block gas ~44M
		maxGas := uint64(44_000_000)
		gasPct := float64(chain.BlockGas) / float64(maxGas)
		gasBar := bar(gasPct, 8)
		lines = append(lines, fmt.Sprintf("%-12s %s / %s %s",
			label("Gas used"),
			fmtGas(chain.BlockGas), fmtGas(maxGas),
			gasBar))
	}

	lines = append(lines, fmt.Sprintf("%-12s %s", label("Gas price"), evm.GasPrice))
	lines = append(lines, "")

	// ERC20 pairs
	lines = append(lines, fmt.Sprintf("%-12s %d registered", label("ERC20 pairs"), len(chain.TokenPairs)))

	// IBC clients
	lines = append(lines, fmt.Sprintf("%-12s %d", label("IBC clients"), chain.IBCClientCount))

	content := joinLines(lines)
	return panelStyle.Width(width).Height(height).Render(content)
}

func fmtGas(g uint64) string {
	if g >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(g)/1_000_000)
	}
	if g >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(g)/1_000)
	}
	return fmt.Sprintf("%d", g)
}
