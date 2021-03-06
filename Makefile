SRC := $(shell find . -name \*.go)


.PHONY: all
all: dep $(SRC)
	go generate
	go install -v

.PHONY: dep
dep:
	go install -v golang.org/x/lint/golint@latest
	go install -v github.com/rakyll/statik@latest

.PHONY: test
test: cover.out

.PHONY: vet
vet: vet.out

.PHONY: lint
lint: lint.out

.PHONY: coverage
coverage: coverage.out

# run the tests with atomic coverage
cover.out: $(SRC)
	go test -v -cover -covermode atomic -coverprofile cover.out ./...

# generate the HTML coverage report
coverage.html: cover.out
	go tool cover -html=cover.out -o coverage.html

# generate the text coverage summary
coverage.out: cover.out
	go tool cover -func=cover.out | tee coverage.out

# run vet against the codebase
vet.out: $(SRC)
	go vet github.com/instana/envcheck/... | tee vet.out

# run the linter against the codebase
lint.out: $(SRC)
	$(shell go list -f {{.Target}} golang.org/x/lint/golint) ./... | tee lint.out

