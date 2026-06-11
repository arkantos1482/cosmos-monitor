package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arkantos1482/cosmos-monitor/internal/fetchall"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/panel"
	"github.com/arkantos1482/cosmos-monitor/internal/render/html"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
)

func main() {
	rpc := flag.String("rpc", "http://localhost:26657", "CometBFT RPC endpoint")
	rest := flag.String("rest", "http://localhost:1317", "Cosmos REST/LCD endpoint")
	evm := flag.String("evm", "http://localhost:8545", "EVM JSON-RPC endpoint")
	container := flag.String("container", "evmd-node", "Docker container name")
	webAddr := flag.String("web", ":7777", "address to serve web UI (e.g. :7777); empty disables")
	dump := flag.Bool("dump", false, "fetch once, print HTML fragment to stdout, and exit")
	flag.Parse()

	load := func(v panel.View) model.Report {
		sn := fetchall.LoadFor(v, *rpc, *rest, *evm, *container)
		return report.Build(sn.Chain, sn.EVM, sn.System, sn.Docker, *evm, sn.Status)
	}

	if *dump {
		rep := load(panel.ViewHome)
		if err := (html.Dump{W: os.Stdout}).Render(rep); err != nil {
			fmt.Fprintf(os.Stderr, "pmtop: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *webAddr == "" {
		fmt.Fprintln(os.Stderr, "pmtop: set -web address or use -dump")
		os.Exit(2)
	}

	html.Start(*webAddr, *evm, load)
}
