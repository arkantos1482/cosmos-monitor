package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/arkantos1482/cosmos-monitor/internal/fetch"
	"github.com/arkantos1482/cosmos-monitor/internal/fetchall"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
	"github.com/arkantos1482/cosmos-monitor/internal/render/html"
	"github.com/arkantos1482/cosmos-monitor/internal/render/terminal"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
	"github.com/arkantos1482/cosmos-monitor/internal/tui"
)

func main() {
	rpc := flag.String("rpc", "http://localhost:26657", "CometBFT RPC endpoint")
	rest := flag.String("rest", "http://localhost:1317", "Cosmos REST/LCD endpoint")
	evm := flag.String("evm", "http://localhost:8545", "EVM JSON-RPC endpoint")
	container := flag.String("container", "evmd-node", "Docker container name")
	webAddr := flag.String("web", "", "address to serve web UI (e.g. :7777); empty = disabled")
	dump := flag.Bool("dump", false, "fetch once, print output, and exit")
	format := flag.String("format", "md", "output format with --dump: md (canonical markdown) or html (fragment)")
	termRender := flag.String("render", "raw", "terminal renderer: raw (canonical markdown) or glamour (styled GFM)")
	flag.Parse()

	if *format != "md" && *format != "html" {
		fmt.Fprintf(os.Stderr, "pmtop: unknown --format %q (use md or html)\n", *format)
		os.Exit(2)
	}
	if *termRender != "raw" && *termRender != "glamour" {
		fmt.Fprintf(os.Stderr, "pmtop: unknown --render %q (use raw or glamour)\n", *termRender)
		os.Exit(2)
	}

	load := func() model.Report {
		sn := fetchall.Load(*rpc, *rest, *evm, *container)
		return report.Build(sn.Chain, sn.EVM, sn.System, sn.Docker, *evm)
	}

	if *dump {
		rep := load()
		var err error
		switch *format {
		case "html":
			err = (html.Dump{W: os.Stdout}).Render(rep)
		default:
			if *termRender == "glamour" {
				err = terminal.Glamour{W: os.Stdout}.Render(rep)
			} else {
				err = terminal.Raw{W: os.Stdout}.Render(rep)
			}
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "pmtop: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *webAddr != "" {
		html.Start(*webAddr, *evm, func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot) {
			sn := fetchall.Load(*rpc, *rest, *evm, *container)
			return sn.Chain, sn.EVM, sn.System, sn.Docker
		})
	}

	if err := tui.Run(tui.Config{
		RPC: *rpc, REST: *rest, EVM: *evm, Container: *container,
		TermRender: *termRender,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "pmtop: %v\n", err)
		os.Exit(1)
	}
}
