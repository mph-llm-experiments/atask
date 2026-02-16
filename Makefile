.PHONY: build test clean

VERSION := 0.1.0
BINARY := atask

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) .

test:
	go test ./...

clean:
	rm -f $(BINARY)

install: build
	cp $(BINARY) ~/bin/