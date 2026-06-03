package main

import (
	"strings"
	"testing"
)

func TestBuildMarkdownWebMermaid(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := buildMarkdown(d, true)
	if !strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("web markdown should embed raw mermaid HTML")
	}
	if !strings.Contains(md, "graph LR") {
		t.Fatal("web economics diagram should use LR layout")
	}
	if strings.Contains(md, "```text") {
		t.Fatal("web markdown should not include ASCII diagram blocks")
	}
}

func TestBuildMarkdownTerminalASCII(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block"}
	md := buildMarkdown(d, false)
	if strings.Contains(md, `<div class="mermaid">`) {
		t.Fatal("terminal markdown should not embed mermaid HTML")
	}
	if !strings.Contains(md, "```text") {
		t.Fatal("terminal markdown should include ASCII diagram blocks")
	}
}

func TestRenderFragmentMermaidHTML(t *testing.T) {
	d := WebData{Inflation: 3.5, PMTEnabled: true, PMTRate: "0.1 PMT/block", GoalBonded: 67}
	out := renderFragment(d)
	if !strings.Contains(out, `class="mermaid"`) {
		t.Fatal("rendered fragment should preserve mermaid div")
	}
	if !strings.Contains(out, "graph LR") {
		t.Fatal("expected LR mermaid source in fragment")
	}
}
