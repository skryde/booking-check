export CGO_ENABLED=0

.PHONY: all
all: dependencies tidy fmt goimports vet staticcheck test

.PHONY: dependencies
dependencies:
	@echo "==> Checking dependencies"
	@type goimports 2> /dev/null > /dev/null || echo "==> Installing 'goimports'" && go install golang.org/x/tools/cmd/goimports@latest > /dev/null
	@type staticcheck 2> /dev/null > /dev/null || echo "==> Installing 'staticcheck'" && go install honnef.co/go/tools/cmd/staticcheck@latest > /dev/null

.PHONY: tidy
tidy:
	@echo "==> Go Mod Tidy"
	@go mod tidy

.PHONY: fmt
fmt:
	@echo "==> Go Fmt"
	@go fmt ./...

.PHONY: goimports
goimports: dependencies
	@echo "==> Goimports"
	@goimports -local "github.com/skryde/booking-check/server" -w .

.PHONY: vet
vet:
	@echo "==> Go Vet"
	@go vet ./...

.PHONY: staticcheck
staticcheck: dependencies
	@echo "==> Staticcheck"
	@staticcheck ./...

.PHONY: test
test:
	@echo "==> Go Test"
	@go test ./... -cover

.PHONY: build
build:
	@echo "==> Go Build"
	@go build -o out/server ./cmd/server/...

.PHONY: clean
clean:
	@echo "==> Clean"
	@rm -Rf out
