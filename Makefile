NODE4_HOST := ec2-34-203-36-91.compute-1.amazonaws.com
KEY        := ~/.ssh/pmt-nodes.pem

.PHONY: push
push: ## git push after committing
	git push

.PHONY: deploy
deploy: ## pull, build, and smoke-test pmtop on node4
	ssh -i $(KEY) ubuntu@$(NODE4_HOST) \
		'cd ~/cosmos-monitor && git pull && /usr/local/go/bin/go build -o ~/pmtop . \
		&& tmux new-session -d -s pmtop -x 220 -y 60 "~/pmtop"; sleep 6; \
		tmux capture-pane -t pmtop -p -S -60; tmux kill-session -t pmtop'

.PHONY: push-deploy
push-deploy: push deploy ## push then deploy to node4

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'
