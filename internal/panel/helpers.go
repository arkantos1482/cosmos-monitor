package panel

import (
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func valLocalMark(v model.Validator) string {
	if v.IsLocal {
		return "**this node**"
	}
	return ""
}

func pmtStatus(d model.Report) string {
	if !d.PMTEnabled {
		return "disabled"
	}
	if d.PMTPoolEmpty {
		return "enabled — pool empty  (no PMT rewards distributing)"
	}
	suffix := ""
	if d.PMTRunway != "" {
		suffix = "  (" + d.PMTRunway + ")"
	}
	return "distributing  " + d.PMTRate + "   pool " + d.PMTBalance + suffix
}
