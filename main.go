package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/arkantos1482/cosmos-monitor/ui"
)

func main() {
	rpc := flag.String("rpc", "http://localhost:26657", "CometBFT RPC endpoint")
	rest := flag.String("rest", "http://localhost:1317", "Cosmos REST/LCD endpoint")
	evm := flag.String("evm", "http://localhost:8545", "EVM JSON-RPC endpoint")
	container := flag.String("container", "evmd-node", "Docker container name")
	flag.Parse()

	cfg := ui.Config{
		RPC:       *rpc,
		REST:      *rest,
		EVM:       *evm,
		Container: *container,
	}

	p := tea.NewProgram(
		ui.New(cfg),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pmtop: %v\n", err)
		os.Exit(1)
	}
}
