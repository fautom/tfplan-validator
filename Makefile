install: mod test
	go install ./cmd/tfplan-validator

test: mod
	gotestsum --format=short-verbose $(TEST) $(TESTARGS)

coverage: mod
	gotestsum --format=short-verbose -- . ./internal/app/tfplan-validator -coverprofile=coverage.txt -covermode=atomic

lint:
	gofmt -s -w .
	misspell .

coverage-html: coverage
	go tool cover -html=coverage.txt

mod:
	go install gotest.tools/gotestsum@latest github.com/client9/misspell/cmd/misspell@latest
	go mod download && go mod verify && go mod tidy

release: mod test lint
	@if [ -z "$${RELEASE}" ]; then echo "ERROR: the RELEASE variable must be specified" && exit 1; fi
	@if [ ! -z "$$(git status --porcelain)" ]; then echo "ERROR: uncommitted changes in repo. check lint" && exit 1; fi
	git tag v${RELEASE}
	git push origin v${RELEASE}

.PHONY: build install test coverage coverage-html mod release
