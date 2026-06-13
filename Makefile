.PHONY: build test install run

build:
	go build -o bin/theirtime ./cmd/theirtime

test:
	go test ./...

install: build
	install -m 755 bin/theirtime $(DESTDIR)/usr/local/bin/theirtime

run: build
	./bin/theirtime $(ARGS)
