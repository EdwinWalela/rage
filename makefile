all: clean test build

build:
	go build -o rage cmd/main.go

test:
	go test -v ./pkg/...

clean:
	rm -rf rage
