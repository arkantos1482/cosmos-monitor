NODE4_HOST := ec2-34-203-36-91.compute-1.amazonaws.com
KEY        := ~/.ssh/pmt-nodes.pem
BINARY     := pmtop
GO_REMOTE  := /usr/local/go/bin/go
REPO_REMOTE := ~/cosmos-monitor
REMOTE_BIN := ~/pmtop

# Remote: stop tmux session and any orphaned ~/pmtop process (nohup/manual starts).
REMOTE_STOP = tmux kill-session -t pmtop 2>/dev/null || true; \
	if pgrep -x pmtop >/dev/null 2>&1; then \
		pkill -x pmtop 2>/dev/null || true; sleep 1; \
		pkill -9 -x pmtop 2>/dev/null || true; \
		echo "stopped pmtop process(es)"; \
	else echo "no pmtop process"; fi

# ── data sources (on node4) ───────────────────────────────────────────────────
# pmtop connects directly on the host (not inside Docker):
#   CometBFT RPC   localhost:26657
#   Cosmos REST    localhost:1317
#   EVM JSON-RPC   localhost:8545
#   Docker socket  /var/run/docker.sock
#
# Source repo: https://github.com/arkantos1482/cosmos-monitor

.DEFAULT_GOAL := help

# ── local dev ─────────────────────────────────────────────────────────────────

.PHONY: build test dump serve
build: ## build ./pmtop from ./cmd/pmtop
	go build -o $(BINARY) ./cmd/pmtop

test: ## run unit tests
	go test ./...

dump: build ## print one-shot HTML fragment to stdout
	./$(BINARY) --dump

serve: build ## run web UI locally on http://localhost:7777
	./$(BINARY)

# ── remote dev (node4) ────────────────────────────────────────────────────────

.PHONY: push deploy push-deploy
push: ## git push after committing
	git push

deploy: ## pull and build on node4 (does not restart running server; use restart)
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'cd $(REPO_REMOTE) && git pull && $(GO_REMOTE) build -o $(REMOTE_BIN) ./cmd/pmtop \
		&& $(REMOTE_BIN) --dump | head -20'

push-deploy: push deploy ## push then deploy to node4

# ── remote ops (node4) ──────────────────────────────────────────────────────────

.PHONY: run start stop restart verify tunnel logs status evmd shell
run: ## run pmtop web UI on node4 foreground (default :7777)
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) '$(REMOTE_BIN)'

start: ## start pmtop in background tmux on node4 (stops any existing instance first)
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'$(REMOTE_STOP); \
		 tmux new-session -d -s pmtop "$(REMOTE_BIN)"; \
		 sleep 1; \
		 pgrep -a pmtop || (echo "failed to start" && exit 1); \
		 tmux capture-pane -t pmtop -p 2>/dev/null | head -3 || true'

stop: ## kill tmux session and any pmtop process on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) '$(REMOTE_STOP)'

restart: ## stop, pull, build, and start pmtop on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'$(REMOTE_STOP); \
		 cd $(REPO_REMOTE) && git pull && $(GO_REMOTE) build -o $(REMOTE_BIN) ./cmd/pmtop && \
		 tmux new-session -d -s pmtop "$(REMOTE_BIN)"; \
		 sleep 1; pgrep -a pmtop'
	@$(MAKE) verify

verify: ## curl node4 UI for fee panel markers (fee-hero)
	@ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'curl -sf http://localhost:7777/ | grep -q fee-hero && echo "OK: fee-hero present" \
		 || (echo "FAIL: old UI still served (run make restart)" && exit 1)'

tunnel: ## SSH tunnel node4 :7777 → localhost:7777 (open http://localhost:7777)
	ssh -i $(KEY) -N -L 7777:localhost:7777 ubuntu@$(NODE4_HOST)

logs: ## tail evmd-node container logs on node4
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker logs -f evmd-node'

status: ## print chain status (RPC + REST) from node4
	@echo "=== CometBFT RPC ==="
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) 'curl -s localhost:26657/status | python3 -m json.tool 2>/dev/null || curl -s localhost:26657/status'
	@echo ""
	@echo "=== Staking validators ==="
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'curl -s "localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED" | python3 -c "import sys,json; vs=json.load(sys.stdin)[\"validators\"]; [print(v[\"description\"][\"moniker\"],v[\"tokens\"]) for v in vs]" 2>/dev/null'

evmd: ## run evmd CLI in validator container (CMD= required)
	@test -n "$(CMD)" || (echo "Usage: CMD=\"query staking validators\" make evmd" && exit 1)
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker exec -it evmd-node evmd $(CMD) --home /data'

shell: ## open shell in evmd-node container on node4
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker exec -it evmd-node /bin/sh'

# ─────────────────────────────────────────────────────────────────────────────

.PHONY: help
help: ## list targets
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
