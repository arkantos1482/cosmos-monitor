package fetch

import "testing"

func TestTracePathLabelHidesValidatorHome(t *testing.T) {
	if got := tracePathLabel("/home/ubuntu/.evmd"); got != "chain-data" {
		t.Fatalf("validator home: %q", got)
	}
	if got := tracePathLabel("/home/ubuntu/.evmd/config/app.toml"); got != "app.toml" {
		t.Fatalf("app.toml: %q", got)
	}
	if got := tracePathLabel("/proc/loadavg"); got != "/proc/loadavg" {
		t.Fatalf("proc: %q", got)
	}
	if got := tracePathLabel("/"); got != "/" {
		t.Fatalf("root: %q", got)
	}
}
