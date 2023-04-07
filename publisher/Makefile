build:
		go test -v ./...
		@go build -o /usr/local/bin/uwp-publisher -ldflags "-s -w"
		@echo "build done: /usr/local/bin/uwp-publisher"

.PHONY: build