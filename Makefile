PHONY: all

all: clean build

clean:
	rm -rf bin

dist: build
	mkdir -p dist
	tar -zcvf dist/magnetron_macos_arm64.tar.gz bin

build:
	go build -o bin/magnetron cmd/magnetron/main.go

run:
	go run cmd/magnetron/main.go

docker:
	docker build -t magnetron:latest .
