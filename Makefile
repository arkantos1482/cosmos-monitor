# Node4 — pmtop dev host today; change NODE4_HOST to point elsewhere.
NODE4_HOST := ec2-34-203-36-91.compute-1.amazonaws.com

# Fleet-wide SSH (same user/key on all validator nodes).
SSH_USER := ubuntu
SSH_KEY  := ~/.ssh/pmt-nodes.pem

# Paths on the remote host (same layout on fleet nodes).
REMOTE_GO    := /usr/local/go/bin/go
REMOTE_REPO  := ~/cosmos-monitor
REMOTE_PMTOP := ~/pmtop

BINARY := pmtop

SSH_NODE4     = ssh -i $(SSH_KEY) $(SSH_USER)@$(NODE4_HOST)
SSH_NODE4_TTY = ssh -t -i $(SSH_KEY) $(SSH_USER)@$(NODE4_HOST)

REMOTE_PMTOP_STOP = tmux kill-session -t pmtop 2>/dev/null || true; \
	if pgrep -x pmtop >/dev/null 2>&1; then \
		pkill -x pmtop 2>/dev/null || true; sleep 1; \
		pkill -9 -x pmtop 2>/dev/null || true; \
		echo "stopped pmtop process(es)"; \
	else echo "no pmtop process"; fi

# On NODE4_HOST (localhost from SSH): CometBFT :26657, REST :1317, EVM :8545, docker socket.
# pmtop source: https://github.com/arkantos1482/cosmos-monitor

.DEFAULT_GOAL := help

# ══════════════════════════════════════════════════════════════════════════════
# Atomic — local
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: build test push tunnel
build: ## atomic local — compile ./pmtop
	go build -o $(BINARY) ./cmd/pmtop

test: ## atomic local — unit tests
	go test ./...

push: ## atomic local — git push
	git push

tunnel: ## atomic local — forward node4 :7777 to localhost:7777
	ssh -i $(SSH_KEY) -N -L 7777:localhost:7777 $(SSH_USER)@$(NODE4_HOST)

# ══════════════════════════════════════════════════════════════════════════════
# Atomic — remote pmtop (on node4)
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: remote-pull remote-build remote-smoke remote-stop remote-start remote-verify remote-run
remote-pull: ## atomic remote pmtop — git pull on node4
	$(SSH_NODE4) 'cd $(REMOTE_REPO) && git pull'

remote-build: ## atomic remote pmtop — go build on node4 (no pull)
	$(SSH_NODE4) 'cd $(REMOTE_REPO) && $(REMOTE_GO) build -o $(REMOTE_PMTOP) ./cmd/pmtop'

remote-smoke: ## atomic remote pmtop — offline --dump on node4 (does not start server)
	$(SSH_NODE4) '$(REMOTE_PMTOP) --dump | head -20'

remote-stop: ## atomic remote pmtop — kill tmux / process on node4
	$(SSH_NODE4) '$(REMOTE_PMTOP_STOP)'

remote-start: ## atomic remote pmtop — start in tmux on node4
	$(SSH_NODE4) \
		'tmux new-session -d -s pmtop "$(REMOTE_PMTOP)"; \
		 sleep 1; \
		 pgrep -a pmtop || (echo "failed to start" && exit 1)'

remote-verify: ## atomic remote pmtop — curl node4 :7777 fee-hero
	@$(SSH_NODE4) \
		'curl -sf http://localhost:7777/ | grep -q fee-hero && echo "OK: fee-hero present" \
		 || (echo "FAIL: fee-hero not found" && exit 1)'

remote-run: ## atomic remote pmtop — foreground on node4 :7777
	$(SSH_NODE4_TTY) '$(REMOTE_PMTOP)'

# ══════════════════════════════════════════════════════════════════════════════
# Atomic — remote validator (on node4, not pmtop)
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: remote-logs remote-status remote-evmd remote-shell
remote-logs: ## atomic remote — tail evmd-node logs on node4
	$(SSH_NODE4_TTY) 'docker logs -f evmd-node'

remote-status: ## atomic remote — CometBFT RPC + validators on node4
	@echo "=== CometBFT RPC ==="
	$(SSH_NODE4) 'curl -s localhost:26657/status | python3 -m json.tool 2>/dev/null || curl -s localhost:26657/status'
	@echo ""
	@echo "=== Staking validators ==="
	$(SSH_NODE4) \
		'curl -s "localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED" | python3 -c "import sys,json; vs=json.load(sys.stdin)[\"validators\"]; [print(v[\"description\"][\"moniker\"],v[\"tokens\"]) for v in vs]" 2>/dev/null'

remote-evmd: ## atomic remote — evmd in container on node4 (CMD= required)
	@test -n "$(CMD)" || (echo "Usage: CMD=\"query staking validators\" make remote-evmd" && exit 1)
	$(SSH_NODE4_TTY) 'docker exec -it evmd-node evmd $(CMD) --home /data'

remote-shell: ## atomic remote — shell in evmd-node on node4
	$(SSH_NODE4_TTY) 'docker exec -it evmd-node /bin/sh'

# ══════════════════════════════════════════════════════════════════════════════
# Integration — compose atomics only (one layer)
# ══════════════════════════════════════════════════════════════════════════════

.PHONY: serve dump remote-deploy remote-restart remote-reload remote-dev-release
serve: build ## integration local — web UI on :7777
	./$(BINARY)

dump: build ## integration local — HTML fragment
	./$(BINARY) --dump

remote-restart: remote-stop remote-start ## integration remote pmtop — recycle server on node4 (no pull/build)

remote-dev-release: push remote-pull remote-build remote-smoke remote-stop remote-start remote-verify ## integration remote — push, then build and run on node4

# ══════════════════════════════════════════════════════════════════════════════

.PHONY: help
help: ## list targets
	@echo "Local:"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*## atomic local ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
	@echo ""
	@echo "Remote pmtop (node4 — $(NODE4_HOST)):"
	@grep -hE '^remote-[a-zA-Z0-9_.-]+:.*## atomic remote pmtop ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
	@echo ""
	@echo "Remote validator (node4 — $(NODE4_HOST)):"
	@grep -hE '^remote-[a-zA-Z0-9_.-]+:.*## atomic remote — ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
	@echo ""
	@echo "Integration:"
	@grep -hE '^[a-zA-Z0-9_.-]+:.*## integration ' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-22s %s\n", $$1, $$2}'
