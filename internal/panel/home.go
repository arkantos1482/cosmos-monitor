package panel

import (
	"fmt"
	"html"
	"strings"

	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeHome(w Writer, d model.Report) {
	w.WriteHTML(`<div class="dash-home">`)
	w.WriteHTML(`<p class="dash-home__lead">Live snapshot — pick a section for full detail. Refreshes every 5s.</p>`)
	w.WriteHTML(`<div class="dash-cards">`)

	syncStr := "synced"
	if !d.Synced {
		syncStr = "CATCHING UP"
	}
	nodeStatus := "stopped"
	if d.NodeRunning {
		nodeStatus = "running"
	}
	pmtStatus := "disabled"
	if d.PMTEnabled {
		pmtStatus = "enabled"
		if d.PMTPoolEmpty {
			pmtStatus = "pool empty"
		}
	}
	govSummary := "no active proposals"
	if n := len(d.Proposals) + len(d.DepositProposals); n > 0 {
		govSummary = fmt.Sprintf("%d proposal(s)", n)
	}
	upgrade := "none"
	if d.UpgradeName != "" && d.UpgradeName != "none" {
		upgrade = fmt.Sprintf("%s @ %s", d.UpgradeName, d.UpgradeHeight)
	}
	evmSync := "synced"
	if !d.EVMSynced {
		evmSync = "syncing"
	}
	localRole := d.Local.SigningStatus
	if d.Local.IsValidator {
		localRole = fmt.Sprintf("%s · %.1f%% VP", d.Local.Status, d.Local.VPPercent)
		if d.Local.Jailed {
			localRole += " · jailed"
		}
	}

	cards := []struct {
		href   string
		title  string
		lines  []string
		badges []struct{ text, kind string }
	}{
		{
			href: "/s/infra", title: "Infrastructure",
			lines: []string{
				fmt.Sprintf("load %.2f / %.2f / %.2f", d.Load1, d.Load5, d.Load15),
				fmt.Sprintf("ram %s / %s (%d%%)", d.MemUsed, d.MemTotal, d.MemPct),
				fmt.Sprintf("disk %d%% · container %s", d.DiskPct, nodeStatus),
			},
			badges: []struct{ text, kind string }{{nodeStatus, badgeKind(nodeStatus)}},
		},
		{
			href: "/s/node", title: "Node",
			lines: []string{
				d.Moniker,
				fmt.Sprintf("height %s · %s", d.BlockHeight, d.TimeSinceBlock),
				fmt.Sprintf("peers %d cosmos · %d evm", d.PeerCount, d.EVMPeerCount),
			},
			badges: []struct{ text, kind string }{{syncStr, badgeKind(syncStr)}},
		},
		{
			href: "/s/validators", title: "Validator set",
			lines: []string{
				fmt.Sprintf("%d bonded", d.BondedCount),
				fmt.Sprintf("%d jailed · %d tombstoned", d.JailedCount, d.TombstonedCount),
			},
		},
		{
			href: "/s/local", title: "This validator",
			lines: []string{localRole},
			badges: localBadges(d),
		},
		{
			href: "/s/economics", title: "Economics",
			lines: []string{
				fmt.Sprintf("bonded %.2f%% · inflation %.2f%%", d.BondedPct, d.Inflation),
				fmt.Sprintf("PMT rewards %s", pmtStatus),
				fmt.Sprintf("base fee %s", d.BaseFee),
			},
			badges: []struct{ text, kind string }{{pmtStatus, badgeKind(pmtStatus)}},
		},
		{
			href: "/s/governance", title: "Governance",
			lines: []string{
				govSummary,
				fmt.Sprintf("upgrade %s", upgrade),
				fmt.Sprintf("IBC clients %d", d.IBCClients),
			},
		},
		{
			href: "/s/evm", title: "EVM JSON-RPC",
			lines: []string{
				fmt.Sprintf("chain %d · block %s", d.EVMChainID, d.EVMBlock),
				fmt.Sprintf("txpool %d pending · %d queued", d.PendingTx, d.QueuedTx),
				fmt.Sprintf("RPC probes %d/%d ok", d.RPCProbeOK, d.RPCProbeTotal),
			},
			badges: []struct{ text, kind string }{{evmSync, badgeKind(evmSync)}},
		},
	}

	for _, c := range cards {
		writeSummaryCard(w, c.href, c.title, c.lines, c.badges)
	}
	w.WriteHTML(`</div></div>`)
}

func localBadges(d model.Report) []struct{ text, kind string } {
	if !d.Local.IsValidator {
		return nil
	}
	var b []struct{ text, kind string }
	if d.Local.Jailed {
		b = append(b, struct{ text, kind string }{"jailed", "bad"})
	}
	if d.Local.Tombstoned {
		b = append(b, struct{ text, kind string }{"tombstoned", "bad"})
	}
	if d.Local.MissedHigh {
		b = append(b, struct{ text, kind string }{"missed blocks high", "warn"})
	}
	return b
}

func badgeKind(v string) string {
	switch badgeClass(v) {
	case "badge--ok":
		return "ok"
	case "badge--warn":
		return "warn"
	case "badge--bad":
		return "bad"
	default:
		return ""
	}
}

func writeSummaryCard(w Writer, href, title string, lines []string, badges []struct{ text, kind string }) {
	view := strings.TrimPrefix(href, "/s/")
	w.WriteHTML(fmt.Sprintf(`<a class="dash-card" href="%s" hx-get="/fragment?view=%s" hx-target="#data" hx-swap="innerHTML scroll:none show:none settle:none" hx-push-url="%s">`,
		html.EscapeString(href), html.EscapeString(view), html.EscapeString(href)))
	w.WriteHTML(fmt.Sprintf(`<h2 class="dash-card__title">%s</h2>`, html.EscapeString(title)))
	if len(badges) > 0 {
		w.WriteHTML(`<div class="dash-card__badges">`)
		for _, b := range badges {
			if b.text == "" {
				continue
			}
			cls := "badge"
			if b.kind != "" {
				cls += " badge--" + b.kind
			}
			w.WriteHTML(fmt.Sprintf(`<span class="%s">%s</span> `, cls, html.EscapeString(b.text)))
		}
		w.WriteHTML(`</div>`)
	}
	w.WriteHTML(`<ul class="dash-card__lines">`)
	for _, line := range lines {
		if line == "" {
			continue
		}
		w.WriteHTML(fmt.Sprintf(`<li>%s</li>`, html.EscapeString(line)))
	}
	w.WriteHTML(`</ul></a>`)
}
