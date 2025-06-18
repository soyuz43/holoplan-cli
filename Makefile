# Makefile

.PHONY: wireframes empty

wireframes:
	holoplan run --stories examples/user_stories.yaml

empty:
	./scripts/empty_output.sh

install:
	pwsh -ExecutionPolicy Bypass -File ./install.ps1