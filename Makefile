PLATFORMS := linux/amd64 linux/arm64 windows/amd64 darwin/arm64 darwin/amd64

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))

release: $(PLATFORMS)

$(PLATFORMS):
	mkdir -p bin/$(os)-$(arch)
	mkdir -p bin/dist
	cd cmd/magnetron && GOOS=$(os) GOARCH=$(arch) go build -o '../../bin/$(os)-$(arch)/magnetron' .
	tar -zcvf 'bin/dist/magnetron-$(os)-$(arch).tar.gz' 'bin/$(os)-$(arch)'

clean:
	rm -rf bin

docker-build:
	docker build -t magnetron:0.5.0 .

docker-run: docker-build
	docker run --rm --name magnetron magnetron:0.5.0

all: clean release $(PLATFORMS)



.PHONY: all
