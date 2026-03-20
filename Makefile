build:
	go build -o kenban .

install: build
	cp kenban /usr/local/bin/kenban
	ln -sf /usr/local/bin/kenban /usr/local/bin/kb

test:
	go test ./...

clean:
	rm -f kenban

.PHONY: build install test clean
