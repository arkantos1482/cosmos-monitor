package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

// RenderStatusStrip returns the global KPI bar HTML for layout embedding.
func RenderStatusStrip(d model.Report) string {
	return renderStatusHTML(d, "")
}

// BuildStatusOOB returns the status bar with HTMX out-of-band swap attributes for poll updates.
func BuildStatusOOB(d model.Report) string {
	return renderStatusHTML(d, `hx-swap-oob="true"`)
}

func renderStatusHTML(d model.Report, extraAttrs string) string {
	var b strings.Builder
	w := newWriter(&b)
	attrs := extraAttrs
	if attrs != "" {
		attrs = " " + attrs
	}
	w.WriteHTML(fmt.Sprintf(`<div id="dash-status" class="dash-status"%s role="region" aria-label="Live status">`, attrs))
	writeStatusPills(w, d)
	w.WriteHTML(`</div>`)
	w.flush()
	return b.String()
}

func writeStatusPills(w Writer, d model.Report) {
	height := "—"
	if d.HasChainStatus {
		height = d.BlockHeight
		if height == "" {
			height = "—"
		}
		if d.TimeSinceBlock != "" {
			height += " · " + d.TimeSinceBlock
		}
	}
	writeStatusPill(w, "Height", height, "")

	if d.HasChainStatus {
		syncStr := "synced"
		syncCls := "badge--ok"
		if !d.Synced {
			syncStr = "catching up"
			syncCls = "badge--warn"
		}
		writeStatusPillBadge(w, "Sync", syncStr, syncCls)
	} else {
		writeStatusPill(w, "Sync", "—", "")
	}

	peers := "—"
	if d.HasChainStatus || d.HasEVMPeers {
		cosmos := "—"
		evm := "—"
		if d.HasChainStatus {
			cosmos = fmt.Sprintf("%d", d.PeerCount)
		}
		if d.HasEVMPeers {
			evm = fmt.Sprintf("%d", d.EVMPeerCount)
		}
		peers = fmt.Sprintf("%s cosmos · %s evm", cosmos, evm)
	}
	writeStatusPill(w, "Peers", peers, "")

	if d.HasNodeStatus {
		nodeStatus := "stopped"
		nodeCls := "badge--bad"
		if d.NodeRunning {
			nodeStatus = "running"
			nodeCls = "badge--ok"
		}
		writeStatusPillBadge(w, "Node", nodeStatus, nodeCls)
	} else {
		writeStatusPill(w, "Node", "—", "")
	}

	baseFee := d.BaseFee
	if !d.HasChainStatus || baseFee == "" {
		baseFee = "—"
	}
	writeStatusPill(w, "Base fee", baseFee, "")

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
	writeStatusPillBadge(w, "PMT", pmtStr, pmtCls)

	refresh := d.TimeUTC
	if refresh == "" {
		refresh = "live"
	}
	w.WriteHTML(fmt.Sprintf(`<time class="dash-status__time">%s</time>`, html.EscapeString(refresh)))
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
