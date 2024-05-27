# .PHONY: default run build test docs clean
# Variables
APP_NAME=rename-files-go

# Tasks
default: run

run:
	@go run ./main.go
build-windows:
	@GOOS=windows GOARCH=amd64 go build -o builds/${APP_NAME}.exe ./main.go
build-mac:
	@go build -o builds/${APP_NAME} ./main.go
build-linux:
	@GOOS=linux GOARCH=amd64 go build -o builds/${APP_NAME}-linux ./main.go