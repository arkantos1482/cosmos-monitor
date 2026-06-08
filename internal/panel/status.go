package panel

import (
	"fmt"
	"html"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeStatusStrip(w Writer, d model.Report) {
	syncStr := "synced"
	syncCls := "badge--ok"
	if !d.Synced {
		syncStr = "catching up"
		syncCls = "badge--warn"
	}

	nodeStatus := "stopped"
	nodeCls := "badge--bad"
	if d.NodeRunning {
		nodeStatus = "running"
		nodeCls = "badge--ok"
	}

	pmtStr := "disabled"
	pmtCls := ""
	if d.PMTEnabled {
		pmtStr = "enabled"
		pmtCls = "badge--ok"
		if d.PMTPoolEmpty {
			pmtStr = "pool empty"
			pmtCls = "badge--warn"
		}
	}

	height := d.BlockHeight
	if height == "" {
		height = "—"
	}
	if d.TimeSinceBlock != "" {
		height += " · " + d.TimeSinceBlock
	}

	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}

	refresh := d.TimeUTC
	if refresh == "" {
		refresh = "live"
	}

	w.WriteHTML(`<div class="dash-status" role="region" aria-label="Live status">`)
	writeStatusPill(w, "Height", height, "")
	writeStatusPillBadge(w, "Sync", syncStr, syncCls)
	writeStatusPill(w, "Peers", fmt.Sprintf("%d cosmos · %d evm", d.PeerCount, d.EVMPeerCount), "")
	writeStatusPillBadge(w, "Node", nodeStatus, nodeCls)
	writeStatusPill(w, "Base fee", baseFee, "")
	writeStatusPillBadge(w, "PMT", pmtStr, pmtCls)
	w.WriteHTML(fmt.Sprintf(`<time class="dash-status__time">%s</time>`, html.EscapeString(refresh)))
	w.WriteHTML(`</div>`)
}

func writeStatusPill(w Writer, label, value, badgeCls string) {
	if badgeCls != "" {
		writeStatusPillBadge(w, label, value, badgeCls)
		return
	}
	w.WriteHTML(fmt.Sprintf(
		`<span class="dash-status__pill"><span class="dash-status__label">%s</span>`+
			`<span class="dash-status__value">%s</span></span>`,
		html.EscapeString(label), html.EscapeString(value),
	))
}

func writeStatusPillBadge(w Writer, label, value, badgeCls string) {
	cls := "badge"
	if badgeCls != "" {
		cls += " " + badgeCls
	}
	w.WriteHTML(fmt.Sprintf(
		`<span class="dash-status__pill"><span class="dash-status__label">%s</span>`+
			`<span class="dash-status__value"><span class="%s">%s</span></span></span>`,
		html.EscapeString(label), cls, html.EscapeString(value),
	))
}
