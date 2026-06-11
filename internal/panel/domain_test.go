package panel

import (
	"strings"
	"testing"
)

func TestEcoBalanceAddrHTML(t *testing.T) {
	out := ecoBalanceAddrHTML("1.00M PMT", "0xABC")
	for _, want := range []string{`class="eco-acct"`, `class="eco-acct__balance"`, `class="eco-acct__addr"`, "1.00M PMT", "0xABC"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in %q", want, out)
		}
	}
}
