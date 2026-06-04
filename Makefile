NODE4_HOST := ec2-34-203-36-91.compute-1.amazonaws.com
KEY        := ~/.ssh/pmt-nodes.pem
BINARY     := pmtop

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

deploy: ## pull, build, and smoke-test HTML dump on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'cd ~/cosmos-monitor && git pull && /usr/local/go/bin/go build -o ~/pmtop ./cmd/pmtop \
		&& ~/pmtop --dump | head -20'

push-deploy: push deploy ## push then deploy to node4

# ── remote ops (node4) ──────────────────────────────────────────────────────────

.PHONY: run start stop tunnel logs status evmd shell
run: ## run pmtop web UI on node4 (default :7777)
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) '~/pmtop'

start: ## start pmtop web UI in background tmux on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'tmux kill-session -t pmtop 2>/dev/null; \
		 tmux new-session -d -s pmtop "~/pmtop"; \
		 sleep 1 && tmux capture-pane -t pmtop -p | head -3'

stop: ## kill background pmtop tmux session on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) 'tmux kill-session -t pmtop 2>/dev/null && echo stopped || echo not running'

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
