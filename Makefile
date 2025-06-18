# Makefile

.PHONY: wireframes empty

wireframes:
	holoplan run --stories examples/user_stories.yaml

empty:
	./scripts/empty_output.sh
