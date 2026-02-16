.PHONY: validate format lint test vulncheck deploy install

validate: format lint test vulncheck

format:
	gofmt -w ./cmd ./internal

lint: format
	golangci-lint run --fix --config=.golangci.yml ./...

test: format
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

vulncheck: format
	govulncheck ./...

deploy:
	mkdir -p ~/.local/bin
	go build -o ~/.local/bin/rv ./cmd/rv

install:
	mkdir -p ~/Projects
	mkdir -p ~/.rivet/worktrees
	$(MAKE) deploy
