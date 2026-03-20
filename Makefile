build:
	go build -o kenban .

install: build
	mkdir -p ~/.local/bin
	cp kenban ~/.local/bin/kenban
	ln -sf ~/.local/bin/kenban ~/.local/bin/kb

test:
	go test ./...

clean:
	rm -f kenban

.PHONY: build install test clean
