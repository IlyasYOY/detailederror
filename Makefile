.DEFAULT_GOAL = verify

BIN=bin
export GOBIN=$(PWD)/$(BIN)

$(BIN)/golangci-lint: 
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0

$(BIN)/gotestsum: 
	go install gotest.tools/gotestsum@v1.11.0

.PHONY: lint 
lint: $(BIN)/golangci-lint
	$(BIN)/golangci-lint run --config=.golangci.yml ./...

.PHONY: test 
test: $(BIN)/gotestsum
	$(BIN)/gotestsum ./... -race

.PHONY: build 
build: 
	go build ./...

.PHONY: test-watch 
test-watch: $(BIN)/gotestsum
	$(BIN)/gotestsum --watch -- ./...

.PHONY: verify 
verify: lint test

