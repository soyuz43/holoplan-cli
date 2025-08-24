# Makefile

.PHONY: wireframes wireframes-figma empty install

# Default: Draw.io format
wireframes:
	holoplan run --stories examples/user_stories.yaml --format drawio

# Figma JSON format
wireframes-figma:
	holoplan run --stories examples/user_stories.yaml --format figma

# Clear output directory
empty:
	./scripts/empty_output.sh

# Install dependencies
install:
	pwsh -ExecutionPolicy Bypass -File ./install.ps1