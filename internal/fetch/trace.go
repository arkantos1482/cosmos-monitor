package fetch

import (
	"sync"
	"time"
)

const maxExchangeResponseBytes = 12_000

// Exchange records one raw data-source call (HTTP, JSON-RPC, file, or fs).
type Exchange struct {
	Kind     string // http, jsonrpc, file, fs, docker
	Method   string
	URL      string
	Request  string
	Response string
	OK       bool
	Error    string
	Latency  time.Duration
}

var (
	traceMu     sync.Mutex
	traceActive bool
	traceLog    []Exchange
)

// BeginTrace starts collecting exchanges until EndTrace.
func BeginTrace() {
	traceMu.Lock()
	traceActive = true
	traceLog = nil
	traceMu.Unlock()
}

// EndTrace returns collected exchanges and clears the collector.
func EndTrace() []Exchange {
	traceMu.Lock()
	defer traceMu.Unlock()
	traceActive = false
	out := append([]Exchange(nil), traceLog...)
	traceLog = nil
	return out
}

func recordTrace(e Exchange) {
	traceMu.Lock()
	if traceActive {
		traceLog = append(traceLog, e)
	}
	traceMu.Unlock()
}

func truncateExchangeResponse(s string) string {
	if len(s) <= maxExchangeResponseBytes {
		return s
	}
	return s[:maxExchangeResponseBytes] + "\n… (truncated)"
}
