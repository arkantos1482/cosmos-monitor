package tui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/arkantos1482/cosmos-monitor/internal/fetchall"
	"github.com/arkantos1482/cosmos-monitor/internal/render/terminal"
	"github.com/arkantos1482/cosmos-monitor/internal/report"
	"golang.org/x/term"
)

// Config holds endpoints for the interactive TUI.
type Config struct {
	RPC, REST, EVM, Container string
}

// lfcrlfWriter translates bare \n to \r\n for terminal output.
type lfcrlfWriter struct{ w io.Writer }

func (t *lfcrlfWriter) Write(p []byte) (int, error) {
	translated := bytes.ReplaceAll(p, []byte("\n"), []byte("\r\n"))
	_, err := t.w.Write(translated)
	return len(p), err
}

// Run starts the interactive refresh loop (r refresh, q quit).
func Run(cfg Config) error {
	out := io.Writer(os.Stdout)
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

	termOut := terminal.Text{W: out}

	refresh := func() {
		fmt.Fprint(out, "\033[H\033[2J")
		fmt.Fprintln(out, "fetching…")
		sn := fetchall.Load(cfg.RPC, cfg.REST, cfg.EVM, cfg.Container)
		rep := report.Build(sn.Chain, sn.EVM, sn.System, sn.Docker, cfg.EVM)
		fmt.Fprint(out, "\033[H\033[2J")
		_ = termOut.Render(rep)
	}

	refresh()

	buf := make([]byte, 1)
	for {
		if _, err := os.Stdin.Read(buf); err != nil {
			return err
		}
		switch buf[0] {
		case 'q', 3:
			return nil
		case 'r':
			refresh()
		}
	}
}
