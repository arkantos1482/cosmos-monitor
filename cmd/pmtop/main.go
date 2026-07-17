package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/arkantos1482/cosmos-monitor/internal/alert"
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
	showSources := flag.Bool("show-sources", false, "show collapsible raw endpoint request/response traces (dev only)")
	alertEnabled := flag.Bool("alert", false, "enable Telegram alerting")
	alertInterval := flag.Duration("alert-interval", 30*time.Second, "alert poll interval")
	alertDryRun := flag.Bool("alert-dry-run", false, "log alert messages without sending to Telegram")
	nodeName := flag.String("node-name", "", "node label in alert messages (default: hostname)")
	flag.Parse()

	opts := panel.Options{ShowSources: showSourcesEnabled(*showSources)}

	load := func(v panel.View) model.Report {
		sn := fetchall.LoadFor(v, *rpc, *rest, *evm, *container)
		return report.Build(sn.Chain, sn.EVM, sn.System, sn.Docker, *evm, sn.Status, sn.AppToml, sn.Exchanges)
	}
	loadHome := func() model.Report { return load(panel.ViewHome) }

	if *dump {
		rep := loadHome()
		if err := (html.Dump{W: os.Stdout, Opts: opts}).Render(rep); err != nil {
			fmt.Fprintf(os.Stderr, "pmtop: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *webAddr == "" && !*alertEnabled {
		fmt.Fprintln(os.Stderr, "pmtop: set -web address, -alert, or use -dump")
		os.Exit(2)
	}

	if *alertEnabled {
		startAlert(loadHome, *nodeName, *alertInterval, *alertDryRun, *webAddr == "")
	}

	if *webAddr != "" {
		html.Start(*webAddr, *evm, load, opts)
	}
}

func startAlert(load func() model.Report, nodeName string, interval time.Duration, dryRun bool, block bool) {
	cfg := alert.LoadConfig(nodeName, interval, dryRun)
	if cfg.NodeName == "" {
		cfg.NodeName, _ = os.Hostname()
	}
	var sender alert.Sender
	if cfg.Enabled() {
		sender = alert.NewTelegramClient(cfg.Token, cfg.ChatID)
	}
	eng := alert.NewEngine(cfg, load, sender)
	if block {
		eng.Run()
		return
	}
	go eng.Run()
}

func showSourcesEnabled(flagVal bool) bool {
	if flagVal {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("PMTOP_SHOW_SOURCES"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
