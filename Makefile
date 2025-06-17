# Makefile

.PHONY: wireframes empty

wireframes:
	go run src/main.go --stories examples/user_stories.yaml

empty:
	./scripts/empty_output.sh
