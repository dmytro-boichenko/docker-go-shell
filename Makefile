.PHONY: build
build:
	@go build -o docker-go-shell ./cmd/$*
