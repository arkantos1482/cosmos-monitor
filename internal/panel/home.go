package panel

import (
	"fmt"
	"html"
	"math"

	"github.com/arkantos1482/cosmos-monitor/internal/feemarket"
	"github.com/arkantos1482/cosmos-monitor/internal/model"
)

func writeHome(w Writer, d model.Report) {
	w.WriteHTML(`<div class="dash-home">`)
	w.WriteHTML(`<p class="dash-home__lead">Live snapshot — pick a section for full detail. Refreshes every 5s.</p>`)

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

	chainCards := []homeCard{
		{
			href: "/s/validators", slug: "validators", title: "Validator set",
			lines: []string{
				fmt.Sprintf("%d bonded", d.BondedCount),
				fmt.Sprintf("%d jailed · %d tombstoned", d.JailedCount, d.TombstonedCount),
			},
		},
		{
			href: "/s/economics", slug: "economics", title: "Economics",
			lines: []string{
				fmt.Sprintf("bonded %.2f%% · inflation %.2f%%", d.BondedPct, d.Inflation),
				fmt.Sprintf("PMT rewards %s", pmtStatus),
			},
			badges: []struct{ text, kind string }{{pmtStatus, badgeKind(pmtStatus)}},
		},
		homeCardFromFeemarket(d),
		{
			href: "/s/governance", slug: "governance", title: "Governance",
			lines: []string{
				govSummary,
				fmt.Sprintf("upgrade %s", upgrade),
				fmt.Sprintf("IBC clients %d", d.IBCClients),
			},
		},
	}
	nodeCards := []homeCard{
		{
			href: "/s/infra", slug: "infra", title: "Infrastructure", span2: true, gauges: true,
			lines: []string{
				fmt.Sprintf("load %.2f / %.2f / %.2f", d.Load1, d.Load5, d.Load15),
				fmt.Sprintf("ram %s / %s (%d%%)", d.MemUsed, d.MemTotal, d.MemPct),
				fmt.Sprintf("disk %d%% · container %s", d.DiskPct, nodeStatus),
			},
			badges: []struct{ text, kind string }{{nodeStatus, badgeKind(nodeStatus)}},
		},
		{
			href: "/s/node", slug: "node", title: "Validator", span2: true, gauges: true,
			lines: validatorCardLines(d, localRole),
			badges: validatorCardBadges(d, syncStr),
		},
		{
			href: "/s/evm", slug: "evm", title: "EVM JSON-RPC",
			lines: []string{
				fmt.Sprintf("chain %d · block %s", d.EVMChainID, d.EVMBlock),
				fmt.Sprintf("txpool %d pending · %d queued", d.PendingTx, d.QueuedTx),
				fmt.Sprintf("RPC probes %d/%d ok", d.RPCProbeOK, d.RPCProbeTotal),
			},
			badges: []struct{ text, kind string }{{evmSync, badgeKind(evmSync)}},
		},
	}

	writeHomeCardGroup(w, NavScopeChain, d, chainCards)
	writeHomeCardGroup(w, NavScopeNode, d, nodeCards)
	w.WriteHTML(`</div>`)
}

type homeCard struct {
	href    string
	slug    string
	title   string
	span2   bool
	gauges  bool
	lines   []string
	badges  []struct{ text, kind string }
}

func homeCardFromFeemarket(d model.Report) homeCard {
	c := feemarketCard(d)
	return homeCard{
		href: c.href, slug: c.slug, title: c.title,
		lines: c.lines, badges: c.badges,
	}
}

func writeHomeCardGroup(w Writer, scope NavScope, d model.Report, cards []homeCard) {
	label := NavScopeLabel(scope)
	if label == "" {
		return
	}
	w.WriteHTML(fmt.Sprintf(`<div class="dash-home__group dash-home__group--%s">`, html.EscapeString(string(scope))))
	w.WriteHTML(fmt.Sprintf(`<h2 class="dash-home__group-title">%s</h2>`, html.EscapeString(label)))
	w.WriteHTML(`<div class="dash-cards dash-cards--bento">`)
	for _, c := range cards {
		writeSummaryCard(w, c.href, c.slug, c.title, c.span2, c.gauges, d, c.lines, c.badges)
	}
	w.WriteHTML(`</div></div>`)
}

func feemarketCard(d model.Report) struct {
	href    string
	slug    string
	title   string
	span2   bool
	gauges  bool
	lines   []string
	badges  []struct{ text, kind string }
} {
	c := feemarket.LoadContext(d)
	baseFee := d.BaseFee
	if baseFee == "" {
		baseFee = "—"
	}
	return struct {
		href    string
		slug    string
		title   string
		span2   bool
		gauges  bool
		lines   []string
		badges  []struct{ text, kind string }
	}{
		href: "/s/feemarket", slug: "feemarket", title: "Fee market",
		lines: []string{
			fmt.Sprintf("base fee %s", baseFee),
			c.HomeCardLine(),
		},
		badges: []struct{ text, kind string }{
			{c.Badge.Label, feemarketBadgeKind(c.Badge)},
		},
	}
}

func feemarketBadgeKind(b feemarket.Badge) string {
	switch b.Class {
	case "rising":
		return "bad"
	case "falling", "floor":
		return "ok"
	case "disabled":
		return "warn"
	default:
		return ""
	}
}

func validatorCardLines(d model.Report, localRole string) []string {
	lines := []string{
		d.Moniker,
		fmt.Sprintf("height %s · %s", d.BlockHeight, d.TimeSinceBlock),
	}
	if d.Local.IsValidator {
		lines = append(lines, localRole)
	}
	lines = append(lines, fmt.Sprintf("peers %d cosmos · %d evm", d.PeerCount, d.EVMPeerCount))
	return lines
}

func validatorCardBadges(d model.Report, syncStr string) []struct{ text, kind string } {
	b := []struct{ text, kind string }{{syncStr, badgeKind(syncStr)}}
	if lb := localBadges(d); len(lb) > 0 {
		b = append(b, lb...)
	}
	return b
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

func writeSummaryCard(w Writer, href, slug, title string, span2, gauges bool, d model.Report, lines []string, badges []struct{ text, kind string }) {
	cls := "dash-card dash-card--" + slug
	if span2 {
		cls += " dash-card--span2"
	}
	w.WriteHTML(fmt.Sprintf(`<a class="%s" href="%s">`, cls, html.EscapeString(href)))
	w.WriteHTML(fmt.Sprintf(`<h2 class="dash-card__title">%s</h2>`, html.EscapeString(title)))
	if len(badges) > 0 {
		w.WriteHTML(`<div class="dash-card__badges">`)
		for _, b := range badges {
			if b.text == "" {
				continue
			}
			bcls := "badge"
			if b.kind != "" {
				bcls += " badge--" + b.kind
			}
			w.WriteHTML(fmt.Sprintf(`<span class="%s">%s</span> `, bcls, html.EscapeString(b.text)))
		}
		w.WriteHTML(`</div>`)
	}
	if gauges {
		writeCardGauges(w, slug, d)
	}
	w.WriteHTML(`<ul class="dash-card__lines">`)
	for _, line := range lines {
		if line == "" {
			continue
		}
		w.WriteHTML(fmt.Sprintf(`<li>%s</li>`, html.EscapeString(line)))
	}
	w.WriteHTML(`</ul>`)
	w.WriteHTML(`<div class="dash-card__footer">View section →</div>`)
	w.WriteHTML(`</a>`)
}

func writeCardGauges(w Writer, slug string, d model.Report) {
	w.WriteHTML(`<div class="dash-card__gauges">`)
	switch slug {
	case "infra":
		writeMiniGauge(w, "RAM", d.MemPct)
		writeMiniGauge(w, "Disk", d.DiskPct)
		loadPct := int(math.Min(d.Load1*100/4, 100)) // normalize load ~4 cores
		if loadPct < 0 {
			loadPct = 0
		}
		writeMiniGauge(w, "Load 1m", loadPct)
	case "node":
		syncPct := 100
		if !d.Synced {
			syncPct = 45
		}
		writeMiniGauge(w, "Sync", syncPct)
	}
	w.WriteHTML(`</div>`)
}

func writeMiniGauge(w Writer, label string, pct int) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	w.WriteHTML(fmt.Sprintf(
		`<div class="mini-gauge"><div class="mini-gauge__label"><span>%s</span><span>%d%%</span></div>`+
			`<div class="mini-gauge__track"><div class="mini-gauge__fill" style="width:%d%%"></div></div></div>`,
		html.EscapeString(label), pct, pct,
	))
}
