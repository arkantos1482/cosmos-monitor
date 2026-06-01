package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/arkantos1482/cosmos-monitor/fetch"
	"golang.org/x/term"
)

// out is the terminal print target; replaced with \r\n-translating writer in raw mode.
var out io.Writer = os.Stdout

// lfcrlfWriter translates bare \n to \r\n, required in raw terminal mode.
type lfcrlfWriter struct{ w io.Writer }

func (t *lfcrlfWriter) Write(p []byte) (int, error) {
	translated := bytes.ReplaceAll(p, []byte("\n"), []byte("\r\n"))
	_, err := t.w.Write(translated)
	return len(p), err
}

func main() {
	rpc       := flag.String("rpc",       "http://localhost:26657", "CometBFT RPC endpoint")
	rest      := flag.String("rest",      "http://localhost:1317",  "Cosmos REST/LCD endpoint")
	evm       := flag.String("evm",       "http://localhost:8545",  "EVM JSON-RPC endpoint")
	container := flag.String("container", "evmd-node",              "Docker container name")
	webAddr   := flag.String("web",       "",                       "address to serve web UI (e.g. :7777); empty = disabled")
	flag.Parse()

	doFetch := func() (fetch.ChainSnapshot, fetch.EVMSnapshot, fetch.SystemSnapshot, fetch.DockerSnapshot) {
		var (
			chain  fetch.ChainSnapshot
			evSnap fetch.EVMSnapshot
			sys    fetch.SystemSnapshot
			docker fetch.DockerSnapshot
			params fetch.ChainParams
			wg     sync.WaitGroup
		)
		wg.Add(5)
		go func() { defer wg.Done(); chain  = fetch.FetchChain(*rpc, *rest) }()
		go func() { defer wg.Done(); evSnap = fetch.FetchEVM(*evm) }()
		go func() { defer wg.Done(); sys    = fetch.FetchSystem() }()
		go func() { defer wg.Done(); docker = fetch.FetchDocker(*container) }()
		go func() { defer wg.Done(); params = fetch.FetchParams(*rest) }()
		wg.Wait()
		chain.Params = params
		return chain, evSnap, sys, docker
	}

	if *webAddr != "" {
		startWeb(*webAddr, doFetch)
	}

	fd := int(os.Stdin.Fd())
	oldState, rawErr := term.MakeRaw(fd)
	restore := func() {
		if rawErr == nil {
			term.Restore(fd, oldState)
		}
	}
	defer restore()

	if rawErr == nil {
		out = &lfcrlfWriter{w: os.Stdout}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		restore()
		os.Exit(0)
	}()

	refresh := func() {
		fmt.Fprint(out, "\033[H\033[2J")
		fmt.Fprintln(out, "fetching…")
		chain, ev, sys, docker := doFetch()
		fmt.Fprint(out, "\033[H\033[2J")
		printDashboard(out, chain, ev, sys, docker)
	}

	refresh()

	buf := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(buf); err != nil {
			return
		}
		switch buf[0] {
		case 'q', 3:
			return
		case 'r':
			refresh()
		}
	}
}
