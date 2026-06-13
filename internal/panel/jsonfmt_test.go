package panel

import (
	"strings"
	"testing"
)

func TestPrettyJSONIndents(t *testing.T) {
	got := prettyJSON(`{"a":1,"b":[2,3]}`, 0)
	if !strings.Contains(got, "\n") {
		t.Fatalf("expected indented JSON, got %q", got)
	}
}

func TestPrettyJSONNonJSONPassthrough(t *testing.T) {
	got := prettyJSON("not json", 0)
	if got != "not json" {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestHighlightJSON(t *testing.T) {
	in := "{\n  \"method\": \"eth_blockNumber\",\n  \"id\": 1,\n  \"ok\": true,\n  \"x\": null\n}"
	out := highlightJSON(in)
	for _, want := range []string{
		`<span class="json-key">`,
		`<span class="json-str">`,
		`<span class="json-num">`,
		`<span class="json-bool">`,
		`<span class="json-null">`,
		`"method"`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}

func TestJSONCodeBlock(t *testing.T) {
	block := jsonCodeBlock(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`, 0)
	if !strings.Contains(block, `json-block`) {
		t.Fatalf("expected json-block wrapper: %s", block)
	}
	if !strings.Contains(block, "\n") {
		t.Fatal("expected pretty-printed JSON inside block")
	}
}
