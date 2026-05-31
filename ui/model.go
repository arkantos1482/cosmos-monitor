package ui

import (
	"github.com/arkantos1482/cosmos-monitor/fetch"
	tea "github.com/charmbracelet/bubbletea"
)

// Config mirrors the top-level config.
type Config struct {
	RPC       string
	REST      string
	EVM       string
	Container string
}

// snapshotMsg carries a fully-fetched snapshot.
type snapshotMsg struct {
	Chain  fetch.ChainSnapshot
	EVM    fetch.EVMSnapshot
	System fetch.SystemSnapshot
	Docker fetch.DockerSnapshot
}

// paramsMsg carries fetched chain params.
type paramsMsg fetch.ChainParams

// errMsg carries a top-level error.
type errMsg struct{ err error }

// Model is the bubbletea model.
type Model struct {
	config       Config
	snapshot     *snapshotMsg
	loading      bool
	params       fetch.ChainParams
	paramsLoaded bool
	width        int
	height       int
	scrollOffset int
}

// New constructs the initial Model.
func New(cfg Config) Model {
	return Model{config: cfg, loading: true}
}

// Init fires the initial fetch.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchParamsCmd(m.config),
		fetchAllCmd(m.config),
	)
}
