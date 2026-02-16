.PHONY: validate format lint test vulncheck deploy install

validate: format lint test vulncheck

format:
	gofmt -w ./cmd ./internal

lint:
	golangci-lint run --config=.golangci.yml ./...

test:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

vulncheck:
	govulncheck ./...

deploy:
	mkdir -p ~/.local/bin
	go build -o ~/.local/bin/rv ./cmd/rv

install:
	mkdir -p ~/Projects
	mkdir -p ~/.rivet/worktrees
	$(MAKE) deploy
