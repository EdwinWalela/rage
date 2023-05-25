all: clean test build

build:
	go build -o rage main.go

test:
	go test -v ./pkg/...

clean:
	rm -rf rage
