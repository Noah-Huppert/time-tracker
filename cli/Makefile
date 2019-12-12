.PHONY: run

BIN ?= timetrk

# run time tracker CLI
run:
	go run .

# build time tracker CLI binary
${BIN}: $(wildcard *.go)
	go build -o ${BIN} .
