package ui

import (
	"sync"
	"time"

	"github.com/arkantos1482/cosmos-monitor/fetch"
	tea "github.com/charmbracelet/bubbletea"
)

func fetchAllCmd(cfg Config) tea.Cmd {
	return func() tea.Msg {
		var (
			chain  fetch.ChainSnapshot
			evm    fetch.EVMSnapshot
			system fetch.SystemSnapshot
			docker fetch.DockerSnapshot
			wg     sync.WaitGroup
		)
		wg.Add(4)
		go func() { defer wg.Done(); chain = fetch.FetchChain(cfg.RPC, cfg.REST) }()
		go func() { defer wg.Done(); evm = fetch.FetchEVM(cfg.EVM) }()
		go func() { defer wg.Done(); system = fetch.FetchSystem() }()
		go func() { defer wg.Done(); docker = fetch.FetchDocker(cfg.Container) }()
		wg.Wait()
		return snapshotMsg{
			Chain:  chain,
			EVM:    evm,
			System: system,
			Docker: docker,
		}
	}
}

func fetchParamsCmd(cfg Config) tea.Cmd {
	return func() tea.Msg {
		p := fetch.FetchParams(cfg.REST)
		return paramsMsg(p)
	}
}

// Update handles messages and key presses.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			return m, fetchAllCmd(m.config)
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j":
			m.scrollOffset++
		case "pgup":
			m.scrollOffset -= m.height / 2
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
		case "pgdown":
			m.scrollOffset += m.height / 2
		case "home", "g":
			m.scrollOffset = 0
		}

	case snapshotMsg:
		s := msg
		s.Chain.Params = m.params
		m.snapshot = &s
		m.loading = false
		return m, nil

	case paramsMsg:
		m.params = fetch.ChainParams(msg)
		m.paramsLoaded = true
		if m.snapshot != nil {
			m.snapshot.Chain.Params = m.params
		}
		return m, nil

	case errMsg:
		// surface error via loading state; snapshot stays nil
		m.loading = false
		return m, nil
	}

	return m, nil
}

// toSnapshot converts snapshotMsg to a unified snapshot.
func (m *Model) toSnapshot() *unifiedSnapshot {
	if m.snapshot == nil {
		return nil
	}
	return &unifiedSnapshot{
		Chain:     m.snapshot.Chain,
		EVM:       m.snapshot.EVM,
		System:    m.snapshot.System,
		Docker:    m.snapshot.Docker,
		FetchedAt: time.Now(),
	}
}

// unifiedSnapshot groups all domain snapshots together.
type unifiedSnapshot struct {
	Chain     fetch.ChainSnapshot
	EVM       fetch.EVMSnapshot
	System    fetch.SystemSnapshot
	Docker    fetch.DockerSnapshot
	FetchedAt time.Time
}
