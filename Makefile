.PHONY: dev

dev:
	@command -v CompileDaemon >/dev/null 2>&1 || (echo "CompileDaemon not found. Run: go install github.com/githubnemo/CompileDaemon@latest" && exit 1)
	@CompileDaemon -build="go build -o ./bin/arkive ./" -command="./bin/arkive"
