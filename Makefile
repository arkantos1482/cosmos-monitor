NODE4_HOST := ec2-34-203-36-91.compute-1.amazonaws.com
KEY        := ~/.ssh/pmt-nodes.pem

# ── data sources ──────────────────────────────────────────────────────────────
# All endpoints are reachable from the node4 host (not inside Docker).
# pmtop connects to these directly; evmd CLI runs inside the evmd-node container.
#
#   CometBFT RPC   localhost:26657
#   Cosmos REST    localhost:1317
#   EVM JSON-RPC   localhost:8545
#   Docker socket  /var/run/docker.sock
#
# Source repo: https://github.com/arkantos1482/cosmos-monitor
# Node4 source: ~/cosmos-monitor/

# ── dev ──────────────────────────────────────────────────────────────────────

# [dev] Commit and push local changes to GitHub.
# Run from tools/ops/pmtop/ after editing source files.
# Prerequisite: changes staged or tracked.
# Usage: make push
.PHONY: push
push: ## git push after committing
	git push

# [dev] Pull latest from GitHub on node4, rebuild pmtop, and smoke-test it.
# Runs the binary in a detached tmux session, captures output, then kills the session.
# Prerequisite: make push (or git push) must have been run first.
# Usage: make deploy
.PHONY: deploy
deploy: ## pull, build, and smoke-test pmtop on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'cd ~/cosmos-monitor && git pull && /usr/local/go/bin/go build -o ~/pmtop . \
		&& tmux new-session -d -s pmtop -x 220 -y 60 "~/pmtop"; sleep 6; \
		tmux capture-pane -t pmtop -p -S -60; tmux kill-session -t pmtop'

# [dev] Push local changes then redeploy to node4 in one step.
# This is the standard iteration loop: edit → commit → make push-deploy.
# Usage: make push-deploy
.PHONY: push-deploy
push-deploy: push deploy ## push then deploy to node4

# ── ops ───────────────────────────────────────────────────────────────────────

# [ops] SSH into node4 and run pmtop interactively (full-screen, press q to quit).
# Use this to inspect live chain state after a deploy or for ad-hoc monitoring.
# Usage: make run
.PHONY: run
run: ## run pmtop interactively on node4
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) '~/pmtop'

# [ops] Tail the validator container logs on node4 (Ctrl-C to stop).
# Use this to watch evmd output, catch panics, or monitor block production.
# Usage: make logs
.PHONY: logs
logs: ## tail evmd-node container logs on node4
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker logs -f evmd-node'

# [ops] Quick chain status check — hits CometBFT RPC and Cosmos REST on node4.
# Prints latest height, sync info, and bonded validators without launching pmtop.
# Usage: make status
.PHONY: status
status: ## print chain status (RPC + REST) from node4
	@echo "=== CometBFT RPC ==="
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) 'curl -s localhost:26657/status | python3 -m json.tool 2>/dev/null || curl -s localhost:26657/status'
	@echo ""
	@echo "=== Staking validators ==="
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'curl -s "localhost:1317/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED" | python3 -c "import sys,json; vs=json.load(sys.stdin)[\"validators\"]; [print(v[\"description\"][\"moniker\"],v[\"tokens\"]) for v in vs]" 2>/dev/null'

# [ops] Run pmtop on node4 with web UI enabled on port 7777.
# Terminal view runs in the foreground (q to quit).
# Open a second terminal and run `make web-tunnel` to reach the browser dashboard.
# Usage: make run-web
.PHONY: run-web
run-web: ## run pmtop + web UI on node4 (port 7777)
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) '~/pmtop --web :7777'

# [ops] SSH tunnel to the pmtop web UI running on node4 (port 7777).
# Prerequisite: pmtop must be running on node4 with --web :7777 (see make run-web).
# Opens http://localhost:7777 in your browser.
# Usage: make web-tunnel
.PHONY: web-tunnel
web-tunnel: ## tunnel pmtop web UI from node4 to localhost:7777
	ssh -i $(KEY) -N -L 7777:localhost:7777 ubuntu@$(NODE4_HOST)

# [ops] Run an evmd CLI command inside the validator container on node4.
# The container home is /data; chain ID is pmt.
# Prerequisite: CMD must be set.
# Usage: CMD="query staking validators --output json" make evmd
#        CMD="query bank balances <address>" make evmd
.PHONY: evmd
evmd: ## run evmd CLI inside the validator container (CMD= required)
	@test -n "$(CMD)" || (echo "Usage: CMD=\"query staking validators\" make evmd" && exit 1)
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker exec -it evmd-node evmd $(CMD) --home /data'

# [ops] Open an interactive shell inside the validator container on node4.
# Use for one-off CLI exploration when you need multiple evmd commands.
# Usage: make shell
.PHONY: shell
shell: ## open a shell inside the evmd-node container on node4
	ssh -t -i $(KEY) ubuntu@$(NODE4_HOST) 'docker exec -it evmd-node /bin/sh'

# ─────────────────────────────────────────────────────────────────────────────

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
