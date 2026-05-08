.PHONY: dev verify test lint deadcode vulncheck gosec

dev:
	@command -v CompileDaemon >/dev/null 2>&1 || (echo "CompileDaemon not found. Run: go install github.com/githubnemo/CompileDaemon@latest" && exit 1)
	@CompileDaemon -build="go build -o ./bin/arkive ./" -command="./bin/arkive" -pattern=".+\\.(go|css|js|svg)$$"

verify: test lint deadcode vulncheck gosec

test:
	@go test ./...

lint:
	@golangci-lint run ./...

deadcode:
	@deadcode .

vulncheck:
	@govulncheck ./...

gosec:
	@gosec ./...
